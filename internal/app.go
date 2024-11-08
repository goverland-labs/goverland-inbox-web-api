package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goverland-labs/goverland-analytics-api-protocol/protobuf/internalapi"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	"github.com/goverland-labs/goverland-inbox-api-protocol/protobuf/inboxapi"
	"github.com/goverland-labs/goverland-platform-events/pkg/natsclient"
	"github.com/nats-io/nats.go"
	"github.com/s-larionov/process-manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/auth"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/config"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/tracking"
	"github.com/goverland-labs/goverland-inbox-web-api/pkg/health"
	"github.com/goverland-labs/goverland-inbox-web-api/pkg/prometheus"
)

type Application struct {
	sigChan <-chan os.Signal
	manager *process.Manager
	cfg     config.App

	feedClient        inboxapi.FeedClient
	achievementClient inboxapi.AchievementClient
	pb                *natsclient.Publisher
}

func NewApplication(cfg config.App) (*Application, error) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	a := &Application{
		sigChan: sigChan,
		cfg:     cfg,
		manager: process.NewManager(),
	}

	err := a.bootstrap()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Application) Run() {
	a.manager.StartAll()
	a.registerShutdown()
}

func (a *Application) bootstrap() error {
	initializers := []func() error{
		// Init Dependencies
		a.initServices,

		// Init Workers: Application
		a.initRESTWorker,

		// Init Workers: System
		a.initPrometheusWorker,
		a.initHealthWorker,
	}

	for _, initializer := range initializers {
		if err := initializer(); err != nil {
			return err
		}
	}

	return nil
}

func (a *Application) initServices() error {
	nc, err := nats.Connect(
		a.cfg.Nats.URL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(a.cfg.Nats.MaxReconnects),
		nats.ReconnectWait(a.cfg.Nats.ReconnectTimeout),
	)
	if err != nil {
		return err
	}

	pb, err := natsclient.NewPublisher(nc)
	if err != nil {
		return err
	}

	a.pb = pb

	return nil
}

func (a *Application) initRESTWorker() error {
	strageConn, err := grpc.Dial(a.cfg.Inbox.StorageAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create connection with storage server: %v", err)
	}

	feedConn, err := grpc.Dial(a.cfg.Inbox.FeedAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create connection with storage server: %v", err)
	}

	anConn, err := grpc.Dial(a.cfg.Analytics.AnalyticsAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create connection with analytics server: %v", err)
	}

	ic := inboxapi.NewUserClient(strageConn)
	pc := inboxapi.NewProposalClient(strageConn)
	sc := inboxapi.NewSubscriptionClient(strageConn)
	settings := inboxapi.NewSettingsClient(strageConn)
	versions := inboxapi.NewAppVersionsClient(strageConn)
	dc := inboxapi.NewDelegateClient(strageConn)
	cs := coresdk.NewClient(a.cfg.Core.CoreURL)
	ac := internalapi.NewAnalyticsClient(anConn)
	a.achievementClient = inboxapi.NewAchievementClient(strageConn)

	a.feedClient = inboxapi.NewFeedClient(feedConn)

	authService := auth.NewService(ic)

	uas := tracking.NewUserActivityService(ic)
	a.manager.AddWorker(process.NewCallbackWorker("user-activity", uas.Start))

	srv, err := rest.NewServer(a.cfg.REST, a.cfg.Chain, authService, cs, sc, settings, versions, a.feedClient, a.achievementClient, ac, ic, pc, dc, uas, a.pb, a.cfg.SiweTTL)
	if err != nil {
		return fmt.Errorf("create REST server: %v", err)
	}
	a.manager.AddWorker(process.NewServerWorker("rest", srv.GetHTTPServer()))

	return nil
}

func (a *Application) initPrometheusWorker() error {
	srv := prometheus.NewServer(a.cfg.Prometheus.Listen, "/metrics")
	a.manager.AddWorker(process.NewServerWorker("prometheus", srv))

	return nil
}

func (a *Application) initHealthWorker() error {
	srv := health.NewHealthCheckServer(a.cfg.Health.Listen, "/status", health.DefaultHandler(a.manager))
	a.manager.AddWorker(process.NewServerWorker("health", srv))

	return nil
}

func (a *Application) registerShutdown() {
	go func(manager *process.Manager) {
		<-a.sigChan

		manager.StopAll()
	}(a.manager)

	a.manager.AwaitAll()
}

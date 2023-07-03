package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	coresdk "github.com/goverland-labs/core-web-sdk"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/s-larionov/process-manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/config"
	"github.com/goverland-labs/inbox-web-api/internal/rest"
	"github.com/goverland-labs/inbox-web-api/pkg/health"
	"github.com/goverland-labs/inbox-web-api/pkg/prometheus"
)

type Application struct {
	sigChan <-chan os.Signal
	manager *process.Manager
	cfg     config.App
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
	// TODO

	return nil
}

func (a *Application) initRESTWorker() error {
	conn, err := grpc.Dial(
		a.cfg.Inbox.StorageAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("create connection with storage server: %v", err)
	}

	ic := inboxapi.NewUserClient(conn)
	sc := inboxapi.NewSubscriptionClient(conn)
	cs := coresdk.NewClient(a.cfg.Core.CoreURL)

	srv := rest.NewServer(a.cfg.REST, auth.NewInMemoryStorage(ic), cs, sc)
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

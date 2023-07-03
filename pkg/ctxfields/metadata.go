package ctxfields

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func EnrichContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, MetadataFieldRequestID, requestID)
}

func EnrichContextWithRequestIP(ctx context.Context, ip string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, MetadataFieldRequestIP, ip)
}

func ExtractRequestIP(ctx context.Context) string {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ""
	}

	ips := md[MetadataFieldRequestIP]
	if len(ips) == 0 {
		return ""
	}

	return ips[0]
}

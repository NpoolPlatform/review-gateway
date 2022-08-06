package api

import (
	"context"

	review "github.com/NpoolPlatform/message/npool/review/gw/v2"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Server struct {
	review.UnimplementedGatewayServer
}

func Register(server grpc.ServiceRegistrar) {
	review.RegisterGatewayServer(server, &Server{})
}

func RegisterGateway(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	if err := review.RegisterGatewayHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
		return err
	}
	return nil
}

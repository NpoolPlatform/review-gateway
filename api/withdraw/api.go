package withdraw

import (
	"context"

	"github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Server struct {
	withdraw.UnimplementedGatewayServer
}

func Register(server grpc.ServiceRegistrar) {
	withdraw.RegisterGatewayServer(server, &Server{})
}

func RegisterGateway(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return withdraw.RegisterGatewayHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
}

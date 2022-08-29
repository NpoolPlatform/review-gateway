package kyc

import (
	"context"

	"github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Server struct {
	kyc.UnimplementedGatewayServer
}

func Register(server grpc.ServiceRegistrar) {
	kyc.RegisterGatewayServer(server, &Server{})
}

func RegisterGateway(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return kyc.RegisterGatewayHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
}

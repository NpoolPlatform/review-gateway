package coupon

import (
	"context"

	coupon1 "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw/coupon"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Server struct {
	coupon1.UnimplementedGatewayServer
}

func Register(server grpc.ServiceRegistrar) {
	coupon1.RegisterGatewayServer(server, &Server{})
}

func RegisterGateway(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return coupon1.RegisterGatewayHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
}

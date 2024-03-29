package api

import (
	"context"

	review "github.com/NpoolPlatform/message/npool/review/gw/v2"

	"github.com/NpoolPlatform/review-gateway/api/kyc"
	"github.com/NpoolPlatform/review-gateway/api/withdraw"
	couponwithdraw "github.com/NpoolPlatform/review-gateway/api/withdraw/coupon"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Server struct {
	review.UnimplementedGatewayServer
}

func Register(server grpc.ServiceRegistrar) {
	review.RegisterGatewayServer(server, &Server{})
	withdraw.Register(server)
	couponwithdraw.Register(server)
	kyc.Register(server)
}

func RegisterGateway(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	if err := review.RegisterGatewayHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
		return err
	}
	if err := withdraw.RegisterGateway(mux, endpoint, opts); err != nil {
		return err
	}
	if err := couponwithdraw.RegisterGateway(mux, endpoint, opts); err != nil {
		return err
	}
	if err := kyc.RegisterGateway(mux, endpoint, opts); err != nil {
		return err
	}
	return nil
}

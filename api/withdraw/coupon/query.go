package coupon

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw/coupon"
	couponwithdraw1 "github.com/NpoolPlatform/review-gateway/pkg/withdraw/coupon"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetCouponWithdrawReviews(ctx context.Context, in *npool.GetCouponWithdrawReviewsRequest) (*npool.GetCouponWithdrawReviewsResponse, error) {
	handler, err := couponwithdraw1.NewHandler(
		ctx,
		couponwithdraw1.WithTargetAppID(&in.AppID, true),
		couponwithdraw1.WithOffset(in.Offset),
		couponwithdraw1.WithLimit(in.Limit),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"GetCouponWithdrawReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetCouponWithdrawReviewsResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	infos, total, err := handler.GetCouponWithdrawReviews(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"GetCouponWithdrawReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetCouponWithdrawReviewsResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.GetCouponWithdrawReviewsResponse{
		Infos: infos,
		Total: total,
	}, nil
}

func (s *Server) GetAppCouponWithdrawReviews(ctx context.Context, in *npool.GetAppCouponWithdrawReviewsRequest) (*npool.GetAppCouponWithdrawReviewsResponse, error) {
	resp, err := s.GetCouponWithdrawReviews(ctx, &npool.GetCouponWithdrawReviewsRequest{
		AppID:  in.TargetAppID,
		Offset: in.Offset,
		Limit:  in.Limit,
	})
	if err != nil {
		logger.Sugar().Errorw(
			"GetAppCouponWithdrawReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetAppCouponWithdrawReviewsResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &npool.GetAppCouponWithdrawReviewsResponse{
		Infos: resp.Infos,
		Total: resp.Total,
	}, nil
}

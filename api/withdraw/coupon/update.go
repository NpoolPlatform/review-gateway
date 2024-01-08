package coupon

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw/coupon"
	couponwithdraw1 "github.com/NpoolPlatform/review-gateway/pkg/withdraw/coupon"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//nolint
func (s *Server) UpdateCouponWithdrawReview(ctx context.Context, in *npool.UpdateCouponWithdrawReviewRequest) (*npool.UpdateCouponWithdrawReviewResponse, error) {
	handler, err := couponwithdraw1.NewHandler(
		ctx,
		couponwithdraw1.WithID(&in.ID, true),
		couponwithdraw1.WithEntID(&in.EntID, true),
		couponwithdraw1.WithAppID(&in.AppID, true),
		couponwithdraw1.WithUserID(&in.UserID, true),
		couponwithdraw1.WithTargetAppID(&in.AppID, true),
		couponwithdraw1.WithState(&in.State, true),
		couponwithdraw1.WithMessage(in.Message, false),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateCouponWithdrawReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateCouponWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := handler.UpdateCouponWithdrawReview(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateCouponWithdrawReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateCouponWithdrawReviewResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.UpdateCouponWithdrawReviewResponse{
		Info: info,
	}, nil
}

//nolint
func (s *Server) UpdateAppCouponWithdrawReview(ctx context.Context, in *npool.UpdateAppCouponWithdrawReviewRequest) (*npool.UpdateAppCouponWithdrawReviewResponse, error) {
	handler, err := couponwithdraw1.NewHandler(
		ctx,
		couponwithdraw1.WithID(&in.ID, true),
		couponwithdraw1.WithEntID(&in.EntID, true),
		couponwithdraw1.WithAppID(&in.AppID, true),
		couponwithdraw1.WithUserID(&in.UserID, true),
		couponwithdraw1.WithTargetAppID(&in.TargetAppID, true),
		couponwithdraw1.WithState(&in.State, true),
		couponwithdraw1.WithMessage(in.Message, false),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateAppCouponWithdrawReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateAppCouponWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := handler.UpdateCouponWithdrawReview(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateAppCouponWithdrawReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateAppCouponWithdrawReviewResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.UpdateAppCouponWithdrawReviewResponse{
		Info: info,
	}, nil
}

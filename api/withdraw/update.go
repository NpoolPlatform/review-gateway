package withdraw

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	withdraw1 "github.com/NpoolPlatform/review-gateway/pkg/withdraw"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateWithdrawReview(ctx context.Context, in *npool.UpdateWithdrawReviewRequest) (*npool.UpdateWithdrawReviewResponse, error) {
	handler, err := withdraw1.NewHandler(
		ctx,
		withdraw1.WithAppID(&in.AppID, true),
		withdraw1.WithUserID(&in.UserID, true),
		withdraw1.WithTargetAppID(&in.AppID, true),
		withdraw1.WithReviewID(&in.ReviewID, true),
		withdraw1.WithState(&in.State, true),
		withdraw1.WithMessage(in.Message, false),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateWithdrawReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := handler.UpdateWithdrawReview(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateWithdrawReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.UpdateWithdrawReviewResponse{
		Info: info,
	}, nil
}

func (s *Server) UpdateAppWithdrawReview(ctx context.Context, in *npool.UpdateAppWithdrawReviewRequest) (*npool.UpdateAppWithdrawReviewResponse, error) {
	handler, err := withdraw1.NewHandler(
		ctx,
		withdraw1.WithAppID(&in.AppID, true),
		withdraw1.WithUserID(&in.UserID, true),
		withdraw1.WithTargetAppID(&in.TargetAppID, true),
		withdraw1.WithReviewID(&in.ReviewID, true),
		withdraw1.WithState(&in.State, true),
		withdraw1.WithMessage(in.Message, false),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateAppWithdrawReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := handler.UpdateWithdrawReview(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateAppWithdrawReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.UpdateAppWithdrawReviewResponse{
		Info: info,
	}, nil
}

package withdraw

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	withdraw1 "github.com/NpoolPlatform/review-gateway/pkg/withdraw"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetWithdrawReviews(ctx context.Context, in *npool.GetWithdrawReviewsRequest) (*npool.GetWithdrawReviewsResponse, error) {
	handler, err := withdraw1.NewHandler(
		ctx,
		withdraw1.WithAppID(&in.AppID),
		withdraw1.WithOffset(in.Offset),
		withdraw1.WithLimit(in.Limit),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"GetWithdrawReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetWithdrawReviewsResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	infos, total, err := handler.GetWithdrawReviews(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"GetWithdrawReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetWithdrawReviewsResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.GetWithdrawReviewsResponse{
		Infos: infos,
		Total: total,
	}, nil
}

func (s *Server) GetAppWithdrawReviews(ctx context.Context, in *npool.GetAppWithdrawReviewsRequest) (*npool.GetAppWithdrawReviewsResponse, error) {
	handler, err := withdraw1.NewHandler(
		ctx,
		withdraw1.WithAppID(&in.TargetAppID),
		withdraw1.WithOffset(in.Offset),
		withdraw1.WithLimit(in.Limit),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"GetAppWithdrawReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetAppWithdrawReviewsResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	infos, total, err := handler.GetWithdrawReviews(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"GetAppWithdrawReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetAppWithdrawReviewsResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.GetAppWithdrawReviewsResponse{
		Infos: infos,
		Total: total,
	}, nil
}

func (s *Server) UpdateWithdrawReview(ctx context.Context, in *npool.UpdateWithdrawReviewRequest) (*npool.UpdateWithdrawReviewResponse, error) {
	handler, err := withdraw1.NewHandler(
		ctx,
		withdraw1.WithAppID(&in.AppID),
		withdraw1.WithUserID(&in.UserID),
		withdraw1.WithReviewID(&in.ReviewID),
		withdraw1.WithState(&in.State, in.Message),
		withdraw1.WithMessage(in.Message),
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
		return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.Internal, "fail update review")
	}

	return &npool.UpdateWithdrawReviewResponse{
		Info: info,
	}, nil
}

func (s *Server) UpdateAppWithdrawReview(ctx context.Context, in *npool.UpdateAppWithdrawReviewRequest) (*npool.UpdateAppWithdrawReviewResponse, error) {
	handler, err := withdraw1.NewHandler(
		ctx,
		withdraw1.WithAppID(&in.AppID),
		withdraw1.WithUserID(&in.UserID),
		withdraw1.WithTargetAppID(&in.TargetAppID),
		withdraw1.WithReviewID(&in.ReviewID),
		withdraw1.WithState(&in.State, in.Message),
		withdraw1.WithMessage(in.Message),
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

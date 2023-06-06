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

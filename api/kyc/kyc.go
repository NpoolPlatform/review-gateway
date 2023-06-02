package kyc

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	kyc1 "github.com/NpoolPlatform/review-gateway/pkg/kyc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetKycReviews(ctx context.Context, in *npool.GetKycReviewsRequest) (*npool.GetKycReviewsResponse, error) {
	handler, err := kyc1.NewHandler(
		ctx,
		kyc1.WithAppID(&in.AppID),
		kyc1.WithOffset(in.Offset),
		kyc1.WithLimit(in.Limit),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"GetKycReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetKycReviewsResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	infos, total, err := handler.GetKycReviews(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"GetKycReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetKycReviewsResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.GetKycReviewsResponse{
		Infos: infos,
		Total: total,
	}, nil
}

func (s *Server) GetAppKycReviews(ctx context.Context, in *npool.GetAppKycReviewsRequest) (*npool.GetAppKycReviewsResponse, error) {
	handler, err := kyc1.NewHandler(
		ctx,
		kyc1.WithAppID(&in.TargetAppID),
		kyc1.WithOffset(in.Offset),
		kyc1.WithLimit(in.Limit),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"GetAppKycReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetAppKycReviewsResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	infos, total, err := handler.GetKycReviews(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"GetAppKycReviews",
			"Req", in,
			"Error", err,
		)
		return &npool.GetAppKycReviewsResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.GetAppKycReviewsResponse{
		Infos: infos,
		Total: total,
	}, nil
}

func (s *Server) UpdateKycReview(ctx context.Context, in *npool.UpdateKycReviewRequest) (*npool.UpdateKycReviewResponse, error) {
	handler, err := kyc1.NewHandler(
		ctx,
		kyc1.WithAppID(&in.AppID),
		kyc1.WithUserID(&in.UserID),
		kyc1.WithLangID(&in.LangID),
		kyc1.WithReviewID(&in.ReviewID),
		kyc1.WithState(&in.State, in.Message),
		kyc1.WithMessage(in.Message),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateKycReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := handler.UpdateKycReview(ctx)
	if err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "error", err)
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.Internal, "fail update review")
	}

	return &npool.UpdateKycReviewResponse{
		Info: info,
	}, nil
}

func (s *Server) UpdateAppKycReview(ctx context.Context, in *npool.UpdateAppKycReviewRequest) (*npool.UpdateAppKycReviewResponse, error) {
	handler, err := kyc1.NewHandler(
		ctx,
		kyc1.WithAppID(&in.AppID),
		kyc1.WithUserID(&in.UserID),
		kyc1.WithLangID(&in.LangID),
		kyc1.WithTargetAppID(&in.TargetAppID),
		kyc1.WithReviewID(&in.ReviewID),
		kyc1.WithState(&in.State, in.Message),
		kyc1.WithMessage(in.Message),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateAppKycReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := handler.UpdateKycReview(ctx)
	if err != nil {
		logger.Sugar().Errorw(
			"UpdateAppKycReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.UpdateAppKycReviewResponse{
		Info: info,
	}, nil
}

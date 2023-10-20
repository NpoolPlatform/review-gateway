package kyc

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	kyc1 "github.com/NpoolPlatform/review-gateway/pkg/kyc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateKycReview(ctx context.Context, in *npool.UpdateKycReviewRequest) (*npool.UpdateKycReviewResponse, error) {
	handler, err := kyc1.NewHandler(
		ctx,
		kyc1.WithAppID(&in.AppID, true),
		kyc1.WithUserID(&in.UserID, true),
		kyc1.WithReviewID(&in.ReviewID, true),
		kyc1.WithState(&in.State, true),
		kyc1.WithMessage(in.Message, false),
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
		logger.Sugar().Errorw(
			"UpdateKycReview",
			"Req", in,
			"Error", err,
		)
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &npool.UpdateKycReviewResponse{
		Info: info,
	}, nil
}

func (s *Server) UpdateAppKycReview(ctx context.Context, in *npool.UpdateAppKycReviewRequest) (*npool.UpdateAppKycReviewResponse, error) {
	handler, err := kyc1.NewHandler(
		ctx,
		kyc1.WithAppID(&in.AppID, true),
		kyc1.WithTargetAppID(&in.TargetAppID, true),
		kyc1.WithUserID(&in.UserID, true),
		kyc1.WithReviewID(&in.ReviewID, true),
		kyc1.WithState(&in.State, true),
		kyc1.WithMessage(in.Message, false),
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

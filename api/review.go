package api

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
)

func (s *Server) UpdateReview(ctx context.Context, in *npool.UpdateReviewRequest) (*npool.UpdateReviewResponse, error) {
	if _, err := uuid.Parse(in.GetID()); err != nil {
		logger.Sugar().Errorw("UpdateReview", "ID", in.GetID(), "error", err)
		return &npool.UpdateReviewResponse{}, status.Error(codes.InvalidArgument, "ID is invalid")
	}

	if _, err := uuid.Parse(in.GetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateReview", "AppID", in.GetAppID(), "error", err)
		return &npool.UpdateReviewResponse{}, status.Error(codes.InvalidArgument, "AppID is invalid")
	}

	if _, err := uuid.Parse(in.GetUserID()); err != nil {
		logger.Sugar().Errorw("UpdateReview", "UserID", in.GetUserID(), "error", err)
		return &npool.UpdateReviewResponse{}, status.Error(codes.InvalidArgument, "UserID is invalid")
	}

	switch in.GetState() {
	case reviewmgrpb.ReviewState_Approved:
	case reviewmgrpb.ReviewState_Rejected:
		if in.GetMessage() == "" {
			logger.Sugar().Errorw("UpdateReview", "State", in.GetState(), "error", "miss rejected reason")
			return &npool.UpdateReviewResponse{}, status.Error(codes.InvalidArgument, "miss rejected reason")
		}
	default:
		logger.Sugar().Errorw("UpdateReview", "State", in.GetState())
		return &npool.UpdateReviewResponse{}, status.Error(codes.InvalidArgument, "State is invalid")
	}

	info, err := review1.UpdateReview(
		ctx,
		in.GetID(), in.GetAppID(), in.GetAppID(), in.GetUserID(),
		in.GetState(), in.GetMessage(),
	)
	if err != nil {
		logger.Sugar().Errorw("UpdateReview", "error", err)
		return &npool.UpdateReviewResponse{}, status.Error(codes.Internal, "fail update review")
	}

	return &npool.UpdateReviewResponse{
		Info: info,
	}, nil
}

func (s *Server) UpdateAppReview(ctx context.Context, in *npool.UpdateAppReviewRequest) (*npool.UpdateAppReviewResponse, error) {
	if _, err := uuid.Parse(in.GetID()); err != nil {
		logger.Sugar().Errorw("UpdateAppReview", "ID", in.GetID(), "error", err)
		return &npool.UpdateAppReviewResponse{}, status.Error(codes.InvalidArgument, "ID is invalid")
	}

	if _, err := uuid.Parse(in.GetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateAppReview", "AppID", in.GetAppID(), "error", err)
		return &npool.UpdateAppReviewResponse{}, status.Error(codes.InvalidArgument, "AppID is invalid")
	}

	if _, err := uuid.Parse(in.GetTargetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateAppReview", "TargetAppID", in.GetTargetAppID(), "error", err)
		return &npool.UpdateAppReviewResponse{}, status.Error(codes.InvalidArgument, "TargetAppID is invalid")
	}

	if _, err := uuid.Parse(in.GetUserID()); err != nil {
		logger.Sugar().Errorw("UpdateAppReview", "UserID", in.GetUserID(), "error", err)
		return &npool.UpdateAppReviewResponse{}, status.Error(codes.InvalidArgument, "UserID is invalid")
	}

	switch in.GetState() {
	case reviewmgrpb.ReviewState_Approved:
	case reviewmgrpb.ReviewState_Rejected:
		if in.GetMessage() == "" {
			logger.Sugar().Errorw("UpdateAppReview", "State", in.GetState(), "error", "miss rejected reason")
			return &npool.UpdateAppReviewResponse{}, status.Error(codes.InvalidArgument, "miss rejected reason")
		}
	default:
		logger.Sugar().Errorw("UpdateAppReview", "State", in.GetState())
		return &npool.UpdateAppReviewResponse{}, status.Error(codes.InvalidArgument, "State is invalid")
	}

	info, err := review1.UpdateReview(
		ctx,
		in.GetID(), in.GetTargetAppID(), in.GetAppID(), in.GetUserID(),
		in.GetState(), in.GetMessage(),
	)
	if err != nil {
		logger.Sugar().Errorw("UpdateAppReview", "error", err)
		return &npool.UpdateAppReviewResponse{}, status.Error(codes.Internal, "fail update review")
	}

	return &npool.UpdateAppReviewResponse{
		Info: info,
	}, nil
}

func (s *Server) GetObjectTypes(ctx context.Context, in *npool.GetObjectTypesRequest) (*npool.GetObjectTypesResponse, error) {
	return &npool.GetObjectTypesResponse{}, status.Error(codes.Unimplemented, "NOT IMPLEMENTED")
}

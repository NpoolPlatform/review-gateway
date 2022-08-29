package kyc

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	constant "github.com/NpoolPlatform/review-gateway/pkg/const"
	kyc1 "github.com/NpoolPlatform/review-gateway/pkg/kyc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
)

func (s *Server) GetKycReviews(
	ctx context.Context, in *npool.GetKycReviewsRequest,
) (
	*npool.GetKycReviewsResponse, error,
) {
	if _, err := uuid.Parse(in.GetAppID()); err != nil {
		logger.Sugar().Errorw("GetKycReviews", "AppID", in.GetAppID(), "error", err)
		return &npool.GetKycReviewsResponse{}, status.Error(codes.InvalidArgument, "AppID is invalid")
	}

	limit := in.GetLimit()
	if limit == 0 {
		limit = constant.DefaultLimitRows
	}

	infos, total, err := kyc1.GetkycReviews(ctx, in.GetAppID(), in.GetOffset(), limit)
	if err != nil {
		logger.Sugar().Errorw("GetKycReviews", "AppID", in.GetAppID(), "error", err)
		return &npool.GetKycReviewsResponse{}, status.Error(codes.InvalidArgument, "fail get withdraw reviews")
	}

	return &npool.GetKycReviewsResponse{
		Infos: infos,
		Total: total,
	}, nil
}

func (s *Server) GetAppKycReviews(
	ctx context.Context, in *npool.GetAppKycReviewsRequest,
) (
	*npool.GetAppKycReviewsResponse, error,
) {
	if _, err := uuid.Parse(in.GetTargetAppID()); err != nil {
		logger.Sugar().Errorw("GetAppKycReviews", "TargetAppID", in.GetTargetAppID(), "error", err)
		return &npool.GetAppKycReviewsResponse{}, status.Error(codes.InvalidArgument, "TargetAppID is invalid")
	}

	resp, err := s.GetKycReviews(ctx, &npool.GetKycReviewsRequest{
		AppID:  in.TargetAppID,
		Offset: in.Offset,
		Limit:  in.Limit,
	})
	if err != nil {
		logger.Sugar().Errorw("GetAppKycReviews", "TargetAppID", in.GetTargetAppID(), "error", err)
		return &npool.GetAppKycReviewsResponse{}, err
	}

	return &npool.GetAppKycReviewsResponse{
		Infos: resp.Infos,
		Total: resp.Total,
	}, nil
}

func (s *Server) UpdateKycReview(
	ctx context.Context, in *npool.UpdateKycReviewRequest,
) (
	*npool.UpdateKycReviewResponse, error,
) {
	if _, err := uuid.Parse(in.GetReviewID()); err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "ID", in.GetReviewID(), "error", err)
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.InvalidArgument, "ReviewID is invalid")
	}

	if _, err := uuid.Parse(in.GetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "AppID", in.GetAppID(), "error", err)
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.InvalidArgument, "AppID is invalid")
	}

	if _, err := uuid.Parse(in.GetUserID()); err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "UserID", in.GetUserID(), "error", err)
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.InvalidArgument, "UserID is invalid")
	}

	if _, err := uuid.Parse(in.GetLangID()); err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "LangID", in.GetLangID(), "error", err)
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.InvalidArgument, "LangID is invalid")
	}

	switch in.GetState() {
	case reviewmgrpb.ReviewState_Approved:
	case reviewmgrpb.ReviewState_Rejected:
		if in.GetMessage() == "" {
			logger.Sugar().Errorw("UpdateKycReview", "State", in.GetState(), "error", "miss rejected reason")
			return &npool.UpdateKycReviewResponse{}, status.Error(codes.InvalidArgument, "miss rejected reason")
		}
	default:
		logger.Sugar().Errorw("UpdateKycReview", "State", in.GetState())
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.InvalidArgument, "State is invalid")
	}

	info, err := kyc1.UpdateKycReview(
		ctx,
		in.GetReviewID(), in.GetAppID(), in.GetAppID(), in.GetUserID(),
		in.GetState(), in.GetMessage(),
	)
	if err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "error", err)
		return &npool.UpdateKycReviewResponse{}, status.Error(codes.Internal, "fail update review")
	}

	return &npool.UpdateKycReviewResponse{
		Info: info,
	}, nil
}

func (s *Server) UpdateAppKycReview(
	ctx context.Context, in *npool.UpdateAppKycReviewRequest,
) (
	*npool.UpdateAppKycReviewResponse, error,
) {
	if _, err := uuid.Parse(in.GetReviewID()); err != nil {
		logger.Sugar().Errorw("UpdateAppKycReview", "ReviewID", in.GetReviewID(), "error", err)
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.InvalidArgument, "ReviewID is invalid")
	}

	if _, err := uuid.Parse(in.GetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateAppKycReview", "AppID", in.GetAppID(), "error", err)
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.InvalidArgument, "AppID is invalid")
	}

	if _, err := uuid.Parse(in.GetTargetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateAppKycReview", "TargetAppID", in.GetTargetAppID(), "error", err)
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.InvalidArgument, "TargetAppID is invalid")
	}

	if _, err := uuid.Parse(in.GetUserID()); err != nil {
		logger.Sugar().Errorw("UpdateAppKycReview", "UserID", in.GetUserID(), "error", err)
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.InvalidArgument, "UserID is invalid")
	}

	if _, err := uuid.Parse(in.GetLangID()); err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "LangID", in.GetLangID(), "error", err)
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.InvalidArgument, "LangID is invalid")
	}

	switch in.GetState() {
	case reviewmgrpb.ReviewState_Approved:
	case reviewmgrpb.ReviewState_Rejected:
		if in.GetMessage() == "" {
			logger.Sugar().Errorw("UpdateAppKycReview", "State", in.GetState(), "error", "miss rejected reason")
			return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.InvalidArgument, "miss rejected reason")
		}
	default:
		logger.Sugar().Errorw("UpdateAppKycReview", "State", in.GetState())
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.InvalidArgument, "State is invalid")
	}

	info, err := kyc1.UpdateKycReview(
		ctx,
		in.GetReviewID(), in.GetTargetAppID(), in.GetAppID(), in.GetUserID(),
		in.GetState(), in.GetMessage(),
	)
	if err != nil {
		logger.Sugar().Errorw("UpdateAppKycReview", "error", err)
		return &npool.UpdateAppKycReviewResponse{}, status.Error(codes.Internal, "fail update review")
	}

	return &npool.UpdateAppKycReviewResponse{
		Info: info,
	}, nil
}

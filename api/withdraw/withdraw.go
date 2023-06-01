package withdraw

import (
	"context"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mw/v2"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	constant "github.com/NpoolPlatform/review-gateway/pkg/const"
	withdraw1 "github.com/NpoolPlatform/review-gateway/pkg/withdraw"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
)

func (s *Server) GetWithdrawReviews(
	ctx context.Context, in *npool.GetWithdrawReviewsRequest,
) (
	*npool.GetWithdrawReviewsResponse, error,
) {
	if _, err := uuid.Parse(in.GetAppID()); err != nil {
		logger.Sugar().Errorw("GetWithdrawReviews", "AppID", in.GetAppID(), "error", err)
		return &npool.GetWithdrawReviewsResponse{}, status.Error(codes.InvalidArgument, "AppID is invalid")
	}

	limit := in.GetLimit()
	if limit == 0 {
		limit = constant.DefaultLimitRows
	}

	infos, total, err := withdraw1.GetWithdrawReviews(ctx, in.GetAppID(), in.GetOffset(), limit)
	if err != nil {
		logger.Sugar().Errorw("GetWithdrawReviews", "AppID", in.GetAppID(), "error", err)
		return &npool.GetWithdrawReviewsResponse{}, status.Error(codes.InvalidArgument, "fail get withdraw reviews")
	}

	return &npool.GetWithdrawReviewsResponse{
		Infos: infos,
		Total: total,
	}, nil
}

func (s *Server) GetAppWithdrawReviews(
	ctx context.Context, in *npool.GetAppWithdrawReviewsRequest,
) (
	*npool.GetAppWithdrawReviewsResponse, error,
) {
	resp, err := s.GetWithdrawReviews(ctx, &npool.GetWithdrawReviewsRequest{
		AppID:  in.TargetAppID,
		Offset: in.Offset,
		Limit:  in.Limit,
	})
	if err != nil {
		logger.Sugar().Errorw("GetWithdrawReviews", "AppID", in.GetTargetAppID(), "error", err)
		return &npool.GetAppWithdrawReviewsResponse{}, err
	}

	return &npool.GetAppWithdrawReviewsResponse{
		Infos: resp.Infos,
		Total: resp.Total,
	}, nil
}

func (s *Server) UpdateWithdrawReview(
	ctx context.Context, in *npool.UpdateWithdrawReviewRequest,
) (
	*npool.UpdateWithdrawReviewResponse, error,
) {
	if _, err := uuid.Parse(in.GetReviewID()); err != nil {
		logger.Sugar().Errorw("UpdateWithdrawReview", "ID", in.GetReviewID(), "error", err)
		return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "ReviewID is invalid")
	}

	if _, err := uuid.Parse(in.GetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateWithdrawReview", "AppID", in.GetAppID(), "error", err)
		return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "AppID is invalid")
	}

	if _, err := uuid.Parse(in.GetUserID()); err != nil {
		logger.Sugar().Errorw("UpdateWithdrawReview", "UserID", in.GetUserID(), "error", err)
		return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "UserID is invalid")
	}

	switch in.GetState() {
	case reviewmgrpb.ReviewState_Approved:
	case reviewmgrpb.ReviewState_Rejected:
		if in.GetMessage() == "" {
			logger.Sugar().Errorw("UpdateWithdrawReview", "State", in.GetState(), "error", "miss rejected reason")
			return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "miss rejected reason")
		}
	default:
		logger.Sugar().Errorw("UpdateWithdrawReview", "State", in.GetState())
		return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "State is invalid")
	}

	info, err := withdraw1.UpdateWithdrawReview(
		ctx,
		in.GetReviewID(), in.GetAppID(), in.GetAppID(), in.GetUserID(),
		in.GetState(), in.Message,
	)
	if err != nil {
		logger.Sugar().Errorw("UpdateWithdrawReview", "error", err)
		return &npool.UpdateWithdrawReviewResponse{}, status.Error(codes.Internal, "fail update review")
	}

	return &npool.UpdateWithdrawReviewResponse{
		Info: info,
	}, nil
}

func (s *Server) UpdateAppWithdrawReview(
	ctx context.Context, in *npool.UpdateAppWithdrawReviewRequest,
) (
	*npool.UpdateAppWithdrawReviewResponse, error,
) {
	if _, err := uuid.Parse(in.GetReviewID()); err != nil {
		logger.Sugar().Errorw("UpdateAppWithdrawReview", "ReviewID", in.GetReviewID(), "error", err)
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "ReviewID is invalid")
	}

	if _, err := uuid.Parse(in.GetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateAppWithdrawReview", "AppID", in.GetAppID(), "error", err)
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "AppID is invalid")
	}

	if _, err := uuid.Parse(in.GetTargetAppID()); err != nil {
		logger.Sugar().Errorw("UpdateAppWithdrawReview", "TargetAppID", in.GetTargetAppID(), "error", err)
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "TargetAppID is invalid")
	}

	if _, err := uuid.Parse(in.GetUserID()); err != nil {
		logger.Sugar().Errorw("UpdateAppWithdrawReview", "UserID", in.GetUserID(), "error", err)
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "UserID is invalid")
	}

	if _, err := uuid.Parse(in.GetLangID()); err != nil {
		logger.Sugar().Errorw("UpdateWithdrawReview", "LangID", in.GetLangID(), "error", err)
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "LangID is invalid")
	}

	switch in.GetState() {
	case reviewmgrpb.ReviewState_Approved:
	case reviewmgrpb.ReviewState_Rejected:
		if in.GetMessage() == "" {
			logger.Sugar().Errorw("UpdateAppWithdrawReview", "State", in.GetState(), "error", "miss rejected reason")
			return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "miss rejected reason")
		}
	default:
		logger.Sugar().Errorw("UpdateAppWithdrawReview", "State", in.GetState())
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.InvalidArgument, "State is invalid")
	}

	info, err := withdraw1.UpdateWithdrawReview(
		ctx,
		in.GetReviewID(), in.GetTargetAppID(), in.GetAppID(), in.GetUserID(),
		in.GetState(), in.Message,
	)
	if err != nil {
		logger.Sugar().Errorw("UpdateAppWithdrawReview", "error", err)
		return &npool.UpdateAppWithdrawReviewResponse{}, status.Error(codes.Internal, "fail update review")
	}

	return &npool.UpdateAppWithdrawReviewResponse{
		Info: info,
	}, nil
}

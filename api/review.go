package api

import (
	"context"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateReview(ctx context.Context, in *npool.UpdateReviewRequest) (*npool.UpdateReviewResponse, error) {
	return &npool.UpdateReviewResponse{}, status.Error(codes.Unimplemented, "NOT IMPLEMENTED")
}

func (s *Server) UpdateAppReview(ctx context.Context, in *npool.UpdateAppReviewRequest) (*npool.UpdateAppReviewResponse, error) {
	return &npool.UpdateAppReviewResponse{}, status.Error(codes.Unimplemented, "NOT IMPLEMENTED")
}

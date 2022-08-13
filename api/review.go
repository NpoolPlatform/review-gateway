package api

import (
	"context"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetObjectTypes(ctx context.Context, in *npool.GetObjectTypesRequest) (*npool.GetObjectTypesResponse, error) {
	return &npool.GetObjectTypesResponse{}, status.Error(codes.Unimplemented, "NOT IMPLEMENTED")
}

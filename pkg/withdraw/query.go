package withdraw

import (
	"context"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
)

func GetWithdrawReviews(ctx context.Context, appID string, offset, limit int32) ([]*npool.WithdrawReview, uint32, error) {
	return nil, 0, nil
}

func GetWithdrawReview(ctx context.Context, reviewID string) (*npool.WithdrawReview, error) {
	return nil, nil
}

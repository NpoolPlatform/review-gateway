package review

import (
	"context"
	"fmt"

	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	npool "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	cli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

func (h *Handler) GetReview(ctx context.Context) (*npool.Review, error) {
	rv, err := cli.GetReview(ctx, *h.ReviewID)
	if err != nil {
		return nil, err
	}
	if rv == nil {
		return nil, fmt.Errorf("invalid review id")
	}

	if rv.State != reviewtypes.ReviewState_Wait {
		return nil, fmt.Errorf("current review state can not be update")
	}

	return rv, nil
}

func (h *Handler) ValidateReview(ctx context.Context) (string, error) {
	rv, err := h.GetReview(ctx)
	if err != nil {
		return "", err
	}
	return rv.ObjectID, nil
}

func (h *Handler) UpdateReview(ctx context.Context) error {
	rv, err := h.GetReview(ctx)
	if err != nil {
		return err
	}

	_, err = cli.UpdateReview(ctx, &npool.ReviewReq{
		ID:         &rv.ID,
		ReviewerID: h.UserID,
		State:      h.State,
		Message:    h.Message,
	})

	return err
}

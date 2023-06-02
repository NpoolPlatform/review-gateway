package review

import (
	"context"
	"fmt"

	npool "github.com/NpoolPlatform/message/npool/review/mw/v2"

	cli "github.com/NpoolPlatform/review-middleware/pkg/client/review"

	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
)

func (h *Handler) ValidateReview(ctx context.Context) (string, error) {
	reviewer, err := usercli.GetUser(ctx, *h.AppID, *h.UserID)
	if err != nil {
		return "", err
	}
	if reviewer == nil {
		return "", fmt.Errorf("invalid reviewer")
	}

	rv, err := cli.GetReview(ctx, *h.ReviewID)
	if err != nil {
		return "", err
	}
	if rv == nil {
		return "", fmt.Errorf("invalid review id")
	}

	if rv.State != npool.ReviewState_Wait {
		return "", fmt.Errorf("invalid review state")
	}

	return rv.ObjectID, nil
}

func (h *Handler) UpdateReview(ctx context.Context) error {
	reviewer, err := usercli.GetUser(ctx, *h.AppID, *h.UserID)
	if err != nil {
		return err
	}
	if reviewer == nil {
		return fmt.Errorf("invalid reviewer")
	}

	rv, err := cli.GetReview(ctx, *h.ReviewID)
	if err != nil {
		return err
	}
	if rv == nil {
		return fmt.Errorf("invalid review id")
	}

	if rv.State != npool.ReviewState_Wait {
		return fmt.Errorf("invalid review state")
	}

	_, err = cli.UpdateReview(ctx, &npool.ReviewReq{
		ID:         &rv.ID,
		ReviewerID: h.UserID,
		State:      h.State,
		Message:    h.Message,
	})

	return err
}

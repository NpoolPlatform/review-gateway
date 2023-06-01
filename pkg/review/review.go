package review

import (
	"context"
	"fmt"

	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mw/v2"

	reviewcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"

	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
)

func ValidateReview(
	ctx context.Context,
	id, appID, reviewerAppID, reviewerID string,
	state reviewmgrpb.ReviewState,
) (string, error) {
	reviewer, err := usercli.GetUser(ctx, reviewerAppID, reviewerID)
	if err != nil {
		return "", err
	}
	if reviewer == nil {
		return "", fmt.Errorf("invalid reviewer")
	}

	switch state {
	case reviewmgrpb.ReviewState_Approved:
	case reviewmgrpb.ReviewState_Rejected:
	default:
		return "", fmt.Errorf("invalid review state")
	}

	rv, err := reviewcli.GetReview(ctx, id)
	if err != nil {
		return "", err
	}
	if rv == nil {
		return "", fmt.Errorf("invalid review id")
	}

	if rv.State != reviewmgrpb.ReviewState_Wait {
		return "", fmt.Errorf("invalid review state")
	}

	return rv.ObjectID, nil
}

func UpdateReview(
	ctx context.Context,
	id, appID, reviewerAppID, reviewerID string,
	state reviewmgrpb.ReviewState,
	message *string,
) error {
	reviewer, err := usercli.GetUser(ctx, reviewerAppID, reviewerID)
	if err != nil {
		return err
	}
	if reviewer == nil {
		return fmt.Errorf("invalid reviewer")
	}

	switch state {
	case reviewmgrpb.ReviewState_Approved:
	case reviewmgrpb.ReviewState_Rejected:
	default:
		return fmt.Errorf("invalid review state")
	}

	rv, err := reviewcli.GetReview(ctx, id)
	if err != nil {
		return err
	}
	if rv == nil {
		return fmt.Errorf("invalid review id")
	}

	if rv.State != reviewmgrpb.ReviewState_Wait {
		return fmt.Errorf("invalid review state")
	}

	_, err = reviewcli.UpdateReview(ctx, &reviewmgrpb.ReviewReq{
		ID:         &rv.ID,
		ReviewerID: &reviewerID,
		State:      &state,
		Message:    message,
	})

	return err
}

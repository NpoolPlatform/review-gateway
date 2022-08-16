package review

import (
	"context"
	"fmt"

	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	reviewcli "github.com/NpoolPlatform/review-service/pkg/client"
	reviewconst "github.com/NpoolPlatform/review-service/pkg/const"

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

	if rv.State != reviewconst.StateWait {
		return "", fmt.Errorf("invalid review state")
	}

	return rv.ObjectID, nil
}

func UpdateReview(
	ctx context.Context,
	id, appID, reviewerAppID, reviewerID string,
	state reviewmgrpb.ReviewState,
	message string,
) error {
	reviewer, err := usercli.GetUser(ctx, reviewerAppID, reviewerID)
	if err != nil {
		return err
	}
	if reviewer == nil {
		return fmt.Errorf("invalid reviewer")
	}

	var rstate string

	switch state {
	case reviewmgrpb.ReviewState_Approved:
		rstate = reviewconst.StateApproved
	case reviewmgrpb.ReviewState_Rejected:
		rstate = reviewconst.StateRejected
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

	if rv.State != reviewconst.StateWait {
		return fmt.Errorf("invalid review state")
	}

	rv.State = rstate
	rv.Message = message
	_, err = reviewcli.UpdateReview(ctx, rv)

	return err
}

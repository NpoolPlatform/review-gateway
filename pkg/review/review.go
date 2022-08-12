package review

import (
	"context"
	"fmt"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	reviewcli "github.com/NpoolPlatform/review-service/pkg/client"
	reviewconst "github.com/NpoolPlatform/review-service/pkg/const"

	withdraw "github.com/NpoolPlatform/review-gateway/pkg/withdraw"

	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
)

func UpdateReview(
	ctx context.Context,
	id, appID, reviewerAppID, reviewerID string,
	state reviewmgrpb.ReviewState,
	message string,
) (
	*npool.Review, error,
) {
	reviewer, err := usercli.GetUser(ctx, reviewerAppID, reviewerID)
	if err != nil {
		return nil, err
	}
	if reviewer == nil {
		return nil, fmt.Errorf("invalid reviewer")
	}

	var stateStr string

	switch state {
	case reviewmgrpb.ReviewState_Approved:
		stateStr = reviewconst.StateApproved
	case reviewmgrpb.ReviewState_Rejected:
		stateStr = reviewconst.StateRejected
	default:
		return nil, fmt.Errorf("invalid review state")
	}

	rv, err := reviewcli.GetReview(ctx, id)
	if err != nil {
		return nil, err
	}
	if rv == nil {
		return nil, fmt.Errorf("invalid review id")
	}

	if rv.State != reviewconst.StateWait {
		return nil, fmt.Errorf("invalid review state")
	}

	var info *npool.Review
	// TODO: move kyc here

	switch rv.ObjectType {
	case reviewmgrpb.ReviewObjectType_ObjectWithdrawal.String():
		info, err = withdraw.UpdateReview(ctx, rv.ObjectID, state)
	default:
		return nil, fmt.Errorf("not supported object type")
	}

	if err != nil {
		return nil, err
	}

	rv.State = stateStr
	rv.Message = message

	rv, err = reviewcli.UpdateReview(ctx, rv)
	if err != nil {
		return nil, err
	}

	info.Message = message
	info.Reviewer = reviewer.EmailAddress
	info.Trigger = reviewmgrpb.ReviewTriggerType(reviewmgrpb.ReviewTriggerType_value[rv.Trigger])
	info.State = state
	info.ObjectType = rv.ObjectType
	info.ObjectID = rv.ObjectID
	info.Domain = rv.Domain
	info.CreatedAt = rv.CreateAt
	info.UpdatedAt = rv.CreateAt // TODO: correct it

	return info, nil
}

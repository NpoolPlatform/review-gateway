package withdraw

import (
	"context"
	"fmt"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"

	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	billingcli "github.com/NpoolPlatform/cloud-hashing-billing/pkg/client"
	withdrawcli "github.com/NpoolPlatform/ledger-manager/pkg/client/withdraw"
	reviewcli "github.com/NpoolPlatform/review-service/pkg/client"
	coininfocli "github.com/NpoolPlatform/sphinx-coininfo/pkg/client"

	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"
)

func GetWithdrawReviews(ctx context.Context, appID string, offset, limit int32) ([]*npool.WithdrawReview, uint32, error) {
	return nil, 0, nil
}

// nolint
func GetWithdrawReview(ctx context.Context, reviewID string) (*npool.WithdrawReview, error) {
	rv, err := reviewcli.GetReview(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	switch rv.ObjectType {
	case "withdraw":
	case reviewmgrpb.ReviewObjectType_ObjectWithdrawal.String():
	default:
		return nil, fmt.Errorf("invalid object type")
	}

	withdraw, err := withdrawcli.GetWithdraw(ctx, rv.ObjectID)
	if err != nil {
		return nil, err
	}
	if withdraw == nil {
		return nil, fmt.Errorf("invalid withdraw")
	}

	user, err := usercli.GetUser(ctx, withdraw.AppID, withdraw.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid user")
	}

	// TODO: we need fill reviewer name, but we miss appid in reviews table
	state := reviewmgrpb.ReviewState_Wait

	switch rv.State {
	case "approved":
		fallthrough // nolint
	case reviewmgrpb.ReviewState_Approved.String():
		state = reviewmgrpb.ReviewState_Approved
	case "rejected":
		fallthrough // nolint
	case reviewmgrpb.ReviewState_Rejected.String():
		state = reviewmgrpb.ReviewState_Rejected
	case "wait":
		fallthrough // nolint
	case reviewmgrpb.ReviewState_Wait.String():
		state = reviewmgrpb.ReviewState_Wait
	default:
		return nil, fmt.Errorf("invalid state")
	}

	trigger := reviewmgrpb.ReviewTriggerType_InsufficientFunds

	switch rv.Trigger {
	case "large amount":
		fallthrough // nolint
	case reviewmgrpb.ReviewTriggerType_LargeAmount.String():
		trigger = reviewmgrpb.ReviewTriggerType_LargeAmount
	case "insufficient":
		fallthrough // nolint
	case reviewmgrpb.ReviewTriggerType_InsufficientFunds.String():
		trigger = reviewmgrpb.ReviewTriggerType_InsufficientFunds
	case "auto review":
		fallthrough // nolint
	case reviewmgrpb.ReviewTriggerType_AutoReviewed.String():
		trigger = reviewmgrpb.ReviewTriggerType_AutoReviewed
	case reviewmgrpb.ReviewTriggerType_InsufficientGas.String():
		trigger = reviewmgrpb.ReviewTriggerType_InsufficientGas
	default:
		return nil, fmt.Errorf("invalid trigger")
	}

	coin, err := coininfocli.GetCoinInfo(ctx, withdraw.CoinTypeID)
	if err != nil {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("invalid coin")
	}

	account, err := billingcli.GetAccount(ctx, withdraw.AccountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("invalid account")
	}

	return &npool.WithdrawReview{
		WithdrawID:            rv.ObjectID,
		ReviewID:              rv.ID,
		UserID:                user.ID,
		KycState:              reviewmgrpb.ReviewState_DefaultReviewState,
		EmailAddress:          user.EmailAddress,
		PhoneNO:               user.PhoneNO,
		Reviewer:              "TODO: to be filled",
		ObjectType:            rv.ObjectType,
		Domain:                rv.Domain,
		CreatedAt:             rv.CreateAt,
		UpdatedAt:             rv.CreateAt,
		Message:               rv.Message,
		State:                 state,
		Trigger:               trigger,
		Amount:                withdraw.Amount,
		FeeAmount:             "TODO: to be filled",
		CoinTypeID:            withdraw.CoinTypeID,
		CoinName:              coin.Name,
		CoinLogo:              coin.Logo,
		CoinUnit:              coin.Unit,
		Address:               account.Address,
		PlatformTransactionID: withdraw.PlatformTransactionID,
		ChainTransactionID:    withdraw.ChainTransactionID,
	}, nil
}

//nolint:dupl
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

	billingconst "github.com/NpoolPlatform/cloud-hashing-billing/pkg/message/const"
	ledgerconst "github.com/NpoolPlatform/ledger-gateway/pkg/message/const"

	userpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	billingpb "github.com/NpoolPlatform/message/npool/cloud-hashing-billing"
	coininfopb "github.com/NpoolPlatform/message/npool/coininfo"
	withdrawmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/withdraw"
	reviewpb "github.com/NpoolPlatform/message/npool/review-service"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	commonpb "github.com/NpoolPlatform/message/npool"

	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
)

// nolint
func GetWithdrawReviews(ctx context.Context, appID string, offset, limit int32) ([]*npool.WithdrawReview, uint32, error) {
	conds := &withdrawmgrpb.Conds{
		AppID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: appID,
		},
	}
	withdraws, total, err := withdrawcli.GetWithdraws(ctx, conds, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	rvs, err := reviewcli.GetDomainReviews(ctx, appID, billingconst.ServiceName, "withdraw")
	if err != nil {
		return nil, 0, err
	}

	rvs1, err := reviewcli.GetDomainReviews(ctx, appID, ledgerconst.ServiceName,
		reviewmgrpb.ReviewObjectType_ObjectWithdrawal.String())
	if err != nil {
		return nil, 0, err
	}

	rvs = append(rvs, rvs1...)

	rvMap := map[string]*reviewpb.Review{}
	for _, rv := range rvs {
		rvMap[rv.ObjectID] = rv
	}

	coins, err := coininfocli.GetCoinInfos(ctx, cruder.NewFilterConds())
	if err != nil {
		return nil, 0, err
	}

	coinMap := map[string]*coininfopb.CoinInfo{}
	for _, coin := range coins {
		coinMap[coin.ID] = coin
	}

	uids := []string{}
	for _, w := range withdraws {
		uids = append(uids, w.UserID)
	}

	users, _, err := usercli.GetManyUsers(ctx, uids)
	if err != nil {
		return nil, 0, err
	}

	userMap := map[string]*userpb.User{}
	for _, user := range users {
		userMap[user.ID] = user
	}

	accounts, err := billingcli.GetAccounts(ctx)
	if err != nil {
		return nil, 0, err
	}

	accMap := map[string]*billingpb.CoinAccountInfo{}
	for _, acc := range accounts {
		accMap[acc.ID] = acc
	}

	infos := []*npool.WithdrawReview{}
	for _, withdraw := range withdraws {
		rv, ok := rvMap[withdraw.ID]
		if !ok {
			return nil, 0, fmt.Errorf("invalid withdraw review")
		}

		coin, ok := coinMap[withdraw.CoinTypeID]
		if !ok {
			return nil, 0, fmt.Errorf("invalid coin")
		}

		user, ok := userMap[withdraw.UserID]
		if !ok {
			return nil, 0, fmt.Errorf("invalid user")
		}

		acc, ok := accMap[withdraw.AccountID]
		if !ok {
			return nil, 0, fmt.Errorf("invalid account")
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
			return nil, 0, fmt.Errorf("invalid state")
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
			return nil, 0, fmt.Errorf("invalid trigger")
		}

		infos = append(infos, &npool.WithdrawReview{
			WithdrawID:            rv.ObjectID,
			WithdrawState:         withdraw.State,
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
			Address:               acc.Address,
			PlatformTransactionID: withdraw.PlatformTransactionID,
			ChainTransactionID:    withdraw.ChainTransactionID,
		})
	}

	return infos, total, nil
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
		WithdrawState:         withdraw.State,
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

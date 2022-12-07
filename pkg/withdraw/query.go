package withdraw

import (
	"context"
	"fmt"

	coininfocli "github.com/NpoolPlatform/chain-middleware/pkg/client/coin"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"

	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	appcoininfocli "github.com/NpoolPlatform/chain-middleware/pkg/client/appcoin"
	withdrawcli "github.com/NpoolPlatform/ledger-manager/pkg/client/withdraw"
	reviewcli "github.com/NpoolPlatform/review-service/pkg/client"

	useraccmwcli "github.com/NpoolPlatform/account-middleware/pkg/client/user"
	useraccmwpb "github.com/NpoolPlatform/message/npool/account/mw/v1/user"

	accountmgrpb "github.com/NpoolPlatform/message/npool/account/mgr/v1/account"

	ledgerconst "github.com/NpoolPlatform/ledger-gateway/pkg/message/const"

	commonpb "github.com/NpoolPlatform/message/npool"
	userpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	appcoinpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/appcoin"
	withdrawmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/withdraw"
	reviewpb "github.com/NpoolPlatform/message/npool/review-service"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

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
	if len(withdraws) == 0 {
		return nil, 0, nil
	}

	rvs, err := reviewcli.GetDomainReviews(ctx, appID, ledgerconst.ServiceName,
		reviewmgrpb.ReviewObjectType_ObjectWithdrawal.String())
	if err != nil {
		return nil, 0, err
	}

	rvMap := map[string]*reviewpb.Review{}
	for _, rv := range rvs {
		rvMap[rv.ObjectID] = rv
	}

	coinTypeIDs := []string{}
	for _, val := range withdraws {
		coinTypeIDs = append(coinTypeIDs, val.CoinTypeID)
	}

	coins, _, err := appcoininfocli.GetCoins(ctx, &appcoinpb.Conds{
		AppID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: appID,
		},
		CoinTypeIDs: &commonpb.StringSliceVal{
			Op:    cruder.IN,
			Value: coinTypeIDs,
		},
	}, 0, int32(len(coinTypeIDs)))
	if err != nil {
		return nil, 0, err
	}

	coinMap := map[string]*appcoinpb.Coin{}
	for _, coin := range coins {
		coinMap[coin.CoinTypeID] = coin
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

	ids := []string{}
	for _, w := range withdraws {
		ids = append(ids, w.AccountID)
	}

	accounts, _, err := useraccmwcli.GetAccounts(ctx, &useraccmwpb.Conds{
		AccountIDs: &commonpb.StringSliceVal{
			Op:    cruder.IN,
			Value: ids,
		},
	}, 0, int32(len(ids)))
	if err != nil {
		return nil, 0, err
	}

	accMap := map[string]*useraccmwpb.Account{}
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

	coin, err := coininfocli.GetCoin(ctx, withdraw.CoinTypeID)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, fmt.Errorf("invalid coin")
	}

	account, err := useraccmwcli.GetAccountOnly(ctx, &useraccmwpb.Conds{
		AppID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.AppID,
		},
		AccountID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.AccountID,
		},
		UsedFor: &commonpb.Int32Val{
			Op:    cruder.EQ,
			Value: int32(accountmgrpb.AccountUsedFor_UserWithdraw),
		},
	})
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

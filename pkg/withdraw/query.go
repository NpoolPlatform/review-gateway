package withdraw

import (
	"context"
	"fmt"

	coininfocli "github.com/NpoolPlatform/chain-middleware/pkg/client/coin"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"

	usermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	appcoinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/app/coin"
	withdrawcli "github.com/NpoolPlatform/ledger-manager/pkg/client/withdraw"
	reviewcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"

	useraccmwcli "github.com/NpoolPlatform/account-middleware/pkg/client/user"
	useraccmwpb "github.com/NpoolPlatform/message/npool/account/mw/v1/user"

	accountmgrpb "github.com/NpoolPlatform/message/npool/account/mgr/v1/account"

	ledgerconst "github.com/NpoolPlatform/ledger-gateway/pkg/message/const"

	commonpb "github.com/NpoolPlatform/message/npool"
	usermwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	appcoinmwpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/app/coin"
	withdrawmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/withdraw"
	reviewpb "github.com/NpoolPlatform/message/npool/review/mw/v2"

	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
)

// nolint
func (h *Handler) GetWithdrawReviews(ctx context.Context) ([]*npool.WithdrawReview, uint32, error) {
	conds := &withdrawmgrpb.Conds{
		AppID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: *h.AppID,
		},
	}
	withdraws, total, err := withdrawcli.GetWithdraws(ctx, conds, h.Offset, h.Limit)
	if err != nil {
		return nil, 0, err
	}
	if len(withdraws) == 0 {
		return nil, 0, nil
	}

	wids := []string{}
	for _, w := range withdraws {
		wids = append(wids, w.ID)
	}

	rvs, err := reviewmwcli.GetObjectReviews(
		ctx,
		*h.AppID,
		ledgerconst.ServiceName,
		wids,
		reviewpb.ReviewObjectType_ObjectWithdrawal,
	)
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

	coins, _, err := appcoinmwcli.GetCoins(ctx, &appcoinmwpb.Conds{
		AppID: &basetypes.StringVal{
			Op:    cruder.EQ,
			Value: *h.AppID,
		},
		CoinTypeIDs: &basetypes.StringSliceVal{
			Op:    cruder.IN,
			Value: coinTypeIDs,
		},
	}, 0, int32(len(coinTypeIDs)))
	if err != nil {
		return nil, 0, err
	}

	coinMap := map[string]*appcoinmwpb.Coin{}
	for _, coin := range coins {
		coinMap[coin.CoinTypeID] = coin
	}

	uids := []string{}
	for _, w := range withdraws {
		uids = append(uids, w.UserID)
	}

	users, _, err := usermwcli.GetUsers(ctx, &usermwpb.Conds{
		IDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: uids},
	}, 0, int32(len(uids)))
	if err != nil {
		return nil, 0, err
	}

	userMap := map[string]*usermwpb.User{}
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
		accMap[acc.AccountID] = acc
	}

	infos := []*npool.WithdrawReview{}
	for _, withdraw := range withdraws {
		rv, ok := rvMap[withdraw.ID]
		if !ok {
			return nil, 0, fmt.Errorf("invalid withdraw review: %v", withdraw)
		}

		coin, ok := coinMap[withdraw.CoinTypeID]
		if !ok {
			continue
		}

		user, ok := userMap[withdraw.UserID]
		if !ok {
			continue
		}

		address := withdraw.Address

		acc, ok := accMap[withdraw.AccountID]
		if ok {
			address = acc.Address
		}

		switch rv.State {
		case reviewpb.ReviewState_Approved:
		case reviewpb.ReviewState_Rejected:
		case reviewpb.ReviewState_Wait:
		default:
			return nil, 0, fmt.Errorf("invalid state")
		}

		switch rv.Trigger {
		case reviewpb.ReviewTriggerType_LargeAmount:
		case reviewpb.ReviewTriggerType_InsufficientFunds:
		case reviewpb.ReviewTriggerType_AutoReviewed:
		case reviewpb.ReviewTriggerType_InsufficientGas:
		case reviewpb.ReviewTriggerType_InsufficientFundsGas:
		default:
			return nil, 0, fmt.Errorf("invalid trigger: %v", rv.Trigger)
		}

		infos = append(infos, &npool.WithdrawReview{
			WithdrawID:            rv.ObjectID,
			WithdrawState:         withdraw.State,
			ReviewID:              rv.ID,
			UserID:                user.ID,
			KycState:              reviewpb.ReviewState_DefaultReviewState,
			EmailAddress:          user.EmailAddress,
			PhoneNO:               user.PhoneNO,
			Reviewer:              "TODO: to be filled",
			ObjectType:            rv.ObjectType,
			Domain:                rv.Domain,
			CreatedAt:             rv.CreatedAt,
			UpdatedAt:             rv.UpdatedAt,
			Message:               rv.Message,
			State:                 rv.State,
			Trigger:               rv.Trigger,
			Amount:                withdraw.Amount,
			FeeAmount:             "TODO: to be filled",
			CoinTypeID:            withdraw.CoinTypeID,
			CoinName:              coin.Name,
			CoinLogo:              coin.Logo,
			CoinUnit:              coin.Unit,
			Address:               address,
			PlatformTransactionID: withdraw.PlatformTransactionID,
			ChainTransactionID:    withdraw.ChainTransactionID,
		})
	}

	return infos, total, nil
}

// nolint
func (h *Handler) GetWithdrawReview(ctx context.Context) (*npool.WithdrawReview, error) {
	rv, err := reviewcli.GetReview(ctx, h.ReviewID.String())
	if err != nil {
		return nil, err
	}

	switch rv.ObjectType {
	case reviewpb.ReviewObjectType_ObjectWithdrawal:
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

	user, err := usermwcli.GetUser(ctx, withdraw.AppID, withdraw.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid user")
	}

	// TODO: we need fill reviewer name, but we miss appid in reviews table

	switch rv.State {
	case reviewpb.ReviewState_Approved:
	case reviewpb.ReviewState_Rejected:
	case reviewpb.ReviewState_Wait:
	default:
		return nil, fmt.Errorf("invalid state")
	}

	switch rv.Trigger {
	case reviewpb.ReviewTriggerType_LargeAmount:
	case reviewpb.ReviewTriggerType_InsufficientFunds:
	case reviewpb.ReviewTriggerType_AutoReviewed:
	case reviewpb.ReviewTriggerType_InsufficientGas:
	case reviewpb.ReviewTriggerType_InsufficientFundsGas:
	default:
		return nil, fmt.Errorf("invalid trigger: %v", rv.Trigger)
	}

	coin, err := coininfocli.GetCoin(ctx, withdraw.CoinTypeID)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, fmt.Errorf("invalid coin")
	}

	address := withdraw.Address

	account, _ := useraccmwcli.GetAccountOnly(ctx, &useraccmwpb.Conds{
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
	if account != nil {
		address = account.Address
	}

	return &npool.WithdrawReview{
		WithdrawID:            rv.ObjectID,
		WithdrawState:         withdraw.State,
		ReviewID:              rv.ID,
		UserID:                user.ID,
		KycState:              reviewpb.ReviewState_DefaultReviewState,
		EmailAddress:          user.EmailAddress,
		PhoneNO:               user.PhoneNO,
		Reviewer:              "TODO: to be filled",
		ObjectType:            rv.ObjectType,
		Domain:                rv.Domain,
		CreatedAt:             rv.CreatedAt,
		UpdatedAt:             rv.UpdatedAt,
		Message:               rv.Message,
		State:                 rv.State,
		Trigger:               rv.Trigger,
		Amount:                withdraw.Amount,
		FeeAmount:             "TODO: to be filled",
		CoinTypeID:            withdraw.CoinTypeID,
		CoinName:              coin.Name,
		CoinLogo:              coin.Logo,
		CoinUnit:              coin.Unit,
		Address:               address,
		PlatformTransactionID: withdraw.PlatformTransactionID,
		ChainTransactionID:    withdraw.ChainTransactionID,
	}, nil
}

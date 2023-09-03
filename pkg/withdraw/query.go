package withdraw

import (
	"context"
	"fmt"

	useraccmwcli "github.com/NpoolPlatform/account-middleware/pkg/client/user"
	usermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	appcoinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/app/coin"
	coinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/coin"
	ledgerconst "github.com/NpoolPlatform/ledger-gateway/pkg/servicename"
	withdrawcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/withdraw"
	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	useraccmwpb "github.com/NpoolPlatform/message/npool/account/mw/v1/user"
	usermwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	appcoinmwpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/app/coin"
	withdrawmwpb "github.com/NpoolPlatform/message/npool/ledger/mw/v2/withdraw"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	reviewmwpb "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

// nolint
func (h *Handler) GetWithdrawReviews(ctx context.Context) ([]*npool.WithdrawReview, uint32, error) {
	conds := &withdrawmwpb.Conds{
		AppID: &basetypes.StringVal{Op: cruder.EQ, Value: *h.AppID},
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
		ledgerconst.ServiceDomain,
		wids,
		reviewtypes.ReviewObjectType_ObjectWithdrawal,
	)
	if err != nil {
		return nil, 0, err
	}

	rvMap := map[string]*reviewmwpb.Review{}
	for _, rv := range rvs {
		rvMap[rv.ObjectID] = rv
	}

	coinTypeIDs := []string{}
	for _, val := range withdraws {
		coinTypeIDs = append(coinTypeIDs, val.CoinTypeID)
	}

	coins, _, err := appcoinmwcli.GetCoins(ctx, &appcoinmwpb.Conds{
		AppID:       &basetypes.StringVal{Op: cruder.EQ, Value: *h.AppID},
		CoinTypeIDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: coinTypeIDs},
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
		AccountIDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
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
		case reviewtypes.ReviewState_Approved:
		case reviewtypes.ReviewState_Rejected:
		case reviewtypes.ReviewState_Wait:
		default:
			return nil, 0, fmt.Errorf("invalid state")
		}

		switch rv.Trigger {
		case reviewtypes.ReviewTriggerType_LargeAmount:
		case reviewtypes.ReviewTriggerType_InsufficientFunds:
		case reviewtypes.ReviewTriggerType_AutoReviewed:
		case reviewtypes.ReviewTriggerType_InsufficientGas:
		case reviewtypes.ReviewTriggerType_InsufficientFundsGas:
		default:
			return nil, 0, fmt.Errorf("invalid trigger: %v", rv.Trigger)
		}

		infos = append(infos, &npool.WithdrawReview{
			WithdrawID:            rv.ObjectID,
			WithdrawState:         withdraw.State,
			ReviewID:              rv.ID,
			UserID:                user.ID,
			KycState:              user.State,
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
	rv, err := reviewmwcli.GetReview(ctx, h.ReviewID.String())
	if err != nil {
		return nil, err
	}

	switch rv.ObjectType {
	case reviewtypes.ReviewObjectType_ObjectWithdrawal:
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
	case reviewtypes.ReviewState_Approved:
	case reviewtypes.ReviewState_Rejected:
	case reviewtypes.ReviewState_Wait:
	default:
		return nil, fmt.Errorf("invalid state")
	}

	switch rv.Trigger {
	case reviewtypes.ReviewTriggerType_LargeAmount:
	case reviewtypes.ReviewTriggerType_InsufficientFunds:
	case reviewtypes.ReviewTriggerType_AutoReviewed:
	case reviewtypes.ReviewTriggerType_InsufficientGas:
	case reviewtypes.ReviewTriggerType_InsufficientFundsGas:
	default:
		return nil, fmt.Errorf("invalid trigger: %v", rv.Trigger)
	}

	coin, err := coinmwcli.GetCoin(ctx, withdraw.CoinTypeID)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, fmt.Errorf("invalid coin")
	}

	address := withdraw.Address

	account, _ := useraccmwcli.GetAccountOnly(ctx, &useraccmwpb.Conds{
		AppID:     &basetypes.StringVal{Op: cruder.EQ, Value: withdraw.AppID},
		AccountID: &basetypes.StringVal{Op: cruder.EQ, Value: withdraw.AccountID},
		UsedFor:   &basetypes.Uint32Val{Op: cruder.EQ, Value: uint32(basetypes.AccountUsedFor_UserWithdraw)},
	})
	if account != nil {
		address = account.Address
	}

	return &npool.WithdrawReview{
		WithdrawID:            rv.ObjectID,
		WithdrawState:         withdraw.State,
		ReviewID:              rv.ID,
		UserID:                user.ID,
		KycState:              user.State,
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

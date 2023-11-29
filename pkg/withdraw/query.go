package withdraw

import (
	"context"
	"fmt"

	useraccmwcli "github.com/NpoolPlatform/account-middleware/pkg/client/user"
	appusermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	appcoinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/app/coin"
	withdrawmwcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/withdraw"
	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	useraccmwpb "github.com/NpoolPlatform/message/npool/account/mw/v1/user"
	appusermwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	appcoinmwpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/app/coin"
	withdrawmwpb "github.com/NpoolPlatform/message/npool/ledger/mw/v2/withdraw"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	reviewmwpb "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

type queryHandler struct {
	*Handler
	userMap    map[string]*appusermwpb.User
	reviewMap  map[string]*reviewmwpb.Review
	coinMap    map[string]*appcoinmwpb.Coin
	accountMap map[string]*useraccmwpb.Account
	withdraws  []*withdrawmwpb.Withdraw
	infos      []*npool.WithdrawReview
}

func (h *queryHandler) getReviews(ctx context.Context) error {
	ids := []string{}
	for _, withdraw := range h.withdraws {
		ids = append(ids, withdraw.ReviewID)
	}

	infos, _, err := reviewmwcli.GetReviews(ctx, &reviewmwpb.Conds{
		ObjectType: &basetypes.Uint32Val{Op: cruder.EQ, Value: uint32(reviewtypes.ReviewObjectType_ObjectWithdrawal)},
		EntIDs:     &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
	}, 0, int32(len(ids)))
	if err != nil {
		return err
	}

	for _, info := range infos {
		h.reviewMap[info.EntID] = info
	}
	return nil
}

func (h *queryHandler) getUsers(ctx context.Context) error {
	ids := []string{}
	for _, withdraw := range h.withdraws {
		ids = append(ids, withdraw.UserID)
	}

	infos, _, err := appusermwcli.GetUsers(ctx, &appusermwpb.Conds{
		IDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
	}, 0, int32(len(ids)))
	if err != nil {
		return err
	}

	for _, info := range infos {
		h.userMap[info.ID] = info
	}
	return nil
}

func (h *queryHandler) getAppCoins(ctx context.Context) error { //nolint
	ids := []string{}
	for _, withdraw := range h.withdraws {
		ids = append(ids, withdraw.CoinTypeID)
	}

	infos, _, err := appcoinmwcli.GetCoins(ctx, &appcoinmwpb.Conds{
		AppID:       &basetypes.StringVal{Op: cruder.EQ, Value: *h.TargetAppID},
		CoinTypeIDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
	}, 0, int32(len(ids)))
	if err != nil {
		return err
	}
	for _, info := range infos {
		h.coinMap[info.CoinTypeID] = info
	}
	return nil
}

func (h *queryHandler) getAccounts(ctx context.Context) error { //nolint
	ids := []string{}
	for _, withdraw := range h.withdraws {
		ids = append(ids, withdraw.AccountID)
	}

	infos, _, err := useraccmwcli.GetAccounts(ctx, &useraccmwpb.Conds{
		AppID:      &basetypes.StringVal{Op: cruder.EQ, Value: *h.TargetAppID},
		AccountIDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
	}, 0, int32(len(ids)))
	if err != nil {
		return err
	}

	for _, info := range infos {
		h.accountMap[info.ID] = info
	}
	return nil
}

func (h *queryHandler) formalize() {
	for _, withdraw := range h.withdraws {
		user, ok := h.userMap[withdraw.UserID]
		if !ok {
			continue
		}
		coin, ok := h.coinMap[withdraw.CoinTypeID]
		if !ok {
			continue
		}
		address := withdraw.Address
		acc, ok := h.accountMap[withdraw.AccountID]
		if ok {
			address = acc.Address
		}
		rv, ok := h.reviewMap[withdraw.ReviewID]
		if !ok {
			continue
		}

		h.infos = append(h.infos, &npool.WithdrawReview{
			UserID:                user.ID,
			KycState:              user.State,
			EmailAddress:          user.EmailAddress,
			PhoneNO:               user.PhoneNO,
			WithdrawID:            withdraw.EntID,
			WithdrawState:         withdraw.State,
			Amount:                withdraw.Amount,
			PlatformTransactionID: withdraw.PlatformTransactionID,
			ChainTransactionID:    withdraw.ChainTransactionID,
			FeeAmount:             "TODO: to be filled",
			CoinTypeID:            withdraw.CoinTypeID,
			CoinName:              coin.Name,
			CoinLogo:              coin.Logo,
			CoinUnit:              coin.Unit,
			Address:               address,
			ReviewID:              rv.EntID,
			Reviewer:              rv.ReviewerID,
			ObjectType:            rv.ObjectType,
			Domain:                rv.Domain,
			CreatedAt:             rv.CreatedAt,
			UpdatedAt:             rv.UpdatedAt,
			Message:               rv.Message,
			State:                 rv.State,
			Trigger:               rv.Trigger,
		})
	}
}

func (h *Handler) GetWithdrawReviews(ctx context.Context) ([]*npool.WithdrawReview, uint32, error) {
	withdraws, total, err := withdrawmwcli.GetWithdraws(ctx, &withdrawmwpb.Conds{
		AppID: &basetypes.StringVal{Op: cruder.EQ, Value: *h.TargetAppID},
	}, h.Offset, h.Limit)
	if err != nil {
		return nil, 0, err
	}
	if len(withdraws) == 0 {
		return nil, 0, nil
	}

	handler := &queryHandler{
		Handler:    h,
		withdraws:  withdraws,
		userMap:    map[string]*appusermwpb.User{},
		coinMap:    map[string]*appcoinmwpb.Coin{},
		reviewMap:  map[string]*reviewmwpb.Review{},
		accountMap: map[string]*useraccmwpb.Account{},
	}

	if err := handler.getReviews(ctx); err != nil {
		return nil, 0, err
	}
	if err := handler.getUsers(ctx); err != nil {
		return nil, 0, err
	}
	if err := handler.getAppCoins(ctx); err != nil {
		return nil, 0, err
	}
	if err := handler.getAccounts(ctx); err != nil {
		return nil, 0, err
	}

	handler.formalize()
	return handler.infos, total, nil
}

// nolint
func (h *Handler) GetWithdrawReview(ctx context.Context) (*npool.WithdrawReview, error) {
	if h.WithdrawID == nil {
		return nil, fmt.Errorf("invalid withdrawid")
	}
	withdraw, err := withdrawmwcli.GetWithdraw(ctx, *h.WithdrawID)
	if err != nil {
		return nil, err
	}
	if withdraw == nil {
		return nil, nil
	}

	handler := &queryHandler{
		Handler:    h,
		withdraws:  []*withdrawmwpb.Withdraw{withdraw},
		userMap:    map[string]*appusermwpb.User{},
		coinMap:    map[string]*appcoinmwpb.Coin{},
		reviewMap:  map[string]*reviewmwpb.Review{},
		accountMap: map[string]*useraccmwpb.Account{},
	}
	if err := handler.getReviews(ctx); err != nil {
		return nil, err
	}
	if err := handler.getUsers(ctx); err != nil {
		return nil, err
	}
	if err := handler.getAppCoins(ctx); err != nil {
		return nil, err
	}
	if err := handler.getAccounts(ctx); err != nil {
		return nil, err
	}

	handler.formalize()
	if len(handler.infos) == 0 {
		return nil, nil
	}
	if len(handler.infos) > 1 {
		return nil, fmt.Errorf("too many record")
	}
	return handler.infos[0], nil

}

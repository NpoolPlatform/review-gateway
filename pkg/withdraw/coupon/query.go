package coupon

import (
	"context"
	"fmt"

	appusermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	appcoinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/app/coin"
	couponmwcli "github.com/NpoolPlatform/inspire-middleware/pkg/client/coupon"
	couponwithdrawmwcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/withdraw/coupon"
	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	appusermwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	appcoinmwpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/app/coin"
	couponmwpb "github.com/NpoolPlatform/message/npool/inspire/mw/v1/coupon"
	couponwithdrawmwpb "github.com/NpoolPlatform/message/npool/ledger/mw/v2/withdraw/coupon"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw/coupon"
	reviewmwpb "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

type queryHandler struct {
	*Handler
	userMap         map[string]*appusermwpb.User
	reviewMap       map[string]*reviewmwpb.Review
	coinMap         map[string]*appcoinmwpb.Coin
	couponMap       map[string]*couponmwpb.Coupon
	couponwithdraws []*couponwithdrawmwpb.CouponWithdraw
	infos           []*npool.CouponWithdrawReview
}

func (h *queryHandler) getReviews(ctx context.Context) error {
	ids := []string{}
	for _, withdraw := range h.couponwithdraws {
		ids = append(ids, withdraw.ReviewID)
	}

	infos, _, err := reviewmwcli.GetReviews(ctx, &reviewmwpb.Conds{
		ObjectType: &basetypes.Uint32Val{Op: cruder.EQ, Value: uint32(reviewtypes.ReviewObjectType_ObjectCouponRandomCash)},
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
	for _, withdraw := range h.couponwithdraws {
		ids = append(ids, withdraw.UserID)
	}

	infos, _, err := appusermwcli.GetUsers(ctx, &appusermwpb.Conds{
		EntIDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
	}, 0, int32(len(ids)))
	if err != nil {
		return err
	}

	for _, info := range infos {
		h.userMap[info.EntID] = info
	}
	return nil
}

func (h *queryHandler) getAppCoins(ctx context.Context) error { //nolint
	ids := []string{}
	for _, withdraw := range h.couponwithdraws {
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

func (h *queryHandler) getCoupons(ctx context.Context) error { //nolint
	ids := []string{}
	for _, withdraw := range h.couponwithdraws {
		ids = append(ids, withdraw.CouponID)
	}

	infos, _, err := couponmwcli.GetCoupons(ctx, &couponmwpb.Conds{
		AppID:  &basetypes.StringVal{Op: cruder.EQ, Value: *h.TargetAppID},
		EntIDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
	}, 0, int32(len(ids)))
	if err != nil {
		return err
	}

	for _, info := range infos {
		h.couponMap[info.EntID] = info
	}
	return nil
}

func (h *queryHandler) formalize() {
	for _, withdraw := range h.couponwithdraws {
		user, ok := h.userMap[withdraw.UserID]
		if !ok {
			continue
		}
		coin, ok := h.coinMap[withdraw.CoinTypeID]
		if !ok {
			continue
		}
		rv, ok := h.reviewMap[withdraw.ReviewID]
		if !ok {
			continue
		}
		coupon, ok := h.couponMap[withdraw.CouponID]
		if !ok {
			continue
		}

		h.infos = append(h.infos, &npool.CouponWithdrawReview{
			ID:                  rv.ID,
			EntID:               rv.EntID,
			UserID:              user.EntID,
			KycState:            user.State,
			EmailAddress:        user.EmailAddress,
			PhoneNO:             user.PhoneNO,
			CouponWithdrawID:    withdraw.EntID,
			CouponWithdrawState: withdraw.State,
			Amount:              withdraw.Amount,
			CoinTypeID:          withdraw.CoinTypeID,
			CoinName:            coin.Name,
			CoinLogo:            coin.Logo,
			CoinUnit:            coin.Unit,
			Reviewer:            rv.ReviewerID,
			ObjectType:          rv.ObjectType,
			Domain:              rv.Domain,
			CreatedAt:           rv.CreatedAt,
			UpdatedAt:           rv.UpdatedAt,
			Message:             rv.Message,
			State:               rv.State,
			Trigger:             rv.Trigger,
			CouponID:            withdraw.CouponID,
			CouponName:          coupon.Message,
		})
	}
}

func (h *Handler) GetCouponWithdrawReviews(ctx context.Context) ([]*npool.CouponWithdrawReview, uint32, error) {
	withdraws, total, err := couponwithdrawmwcli.GetCouponWithdraws(ctx, &couponwithdrawmwpb.Conds{
		AppID: &basetypes.StringVal{Op: cruder.EQ, Value: *h.TargetAppID},
	}, h.Offset, h.Limit)
	if err != nil {
		return nil, 0, err
	}
	if len(withdraws) == 0 {
		return nil, 0, nil
	}

	handler := &queryHandler{
		Handler:         h,
		couponwithdraws: withdraws,
		userMap:         map[string]*appusermwpb.User{},
		coinMap:         map[string]*appcoinmwpb.Coin{},
		reviewMap:       map[string]*reviewmwpb.Review{},
		couponMap:       map[string]*couponmwpb.Coupon{},
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
	if err := handler.getCoupons(ctx); err != nil {
		return nil, 0, err
	}

	handler.formalize()
	return handler.infos, total, nil
}

// nolint
func (h *Handler) GetCouponWithdrawReview(ctx context.Context) (*npool.CouponWithdrawReview, error) {
	if h.CouponWithdrawID == nil {
		return nil, fmt.Errorf("invalid withdrawid")
	}
	withdraw, err := couponwithdrawmwcli.GetCouponWithdraw(ctx, *h.CouponWithdrawID)
	if err != nil {
		return nil, err
	}
	if withdraw == nil {
		return nil, nil
	}

	handler := &queryHandler{
		Handler:         h,
		couponwithdraws: []*couponwithdrawmwpb.CouponWithdraw{withdraw},
		userMap:         map[string]*appusermwpb.User{},
		coinMap:         map[string]*appcoinmwpb.Coin{},
		reviewMap:       map[string]*reviewmwpb.Review{},
		couponMap:       map[string]*couponmwpb.Coupon{},
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
	if err := handler.getCoupons(ctx); err != nil {
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

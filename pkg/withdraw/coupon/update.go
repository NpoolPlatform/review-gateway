package coupon

import (
	"context"
	"fmt"

	kycmwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/kyc"
	usermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	coinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/coin"
	couponwithdrawmwcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/withdraw/coupon"
	"github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	kycmwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/kyc"
	ledgertypes "github.com/NpoolPlatform/message/npool/basetypes/ledger/v1"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	couponwithdrawmwpb "github.com/NpoolPlatform/message/npool/ledger/mw/v2/withdraw/coupon"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw/coupon"
	reviewmwpb "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

type updateHandler struct {
	*Handler
	review         *reviewmwpb.Review
	couponwithdraw *couponwithdrawmwpb.CouponWithdraw
}

func (h *updateHandler) checkUser(ctx context.Context) error {
	info, err := usermwcli.GetUser(ctx, *h.AppID, *h.UserID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("invalid user")
	}
	return nil
}

func (h *updateHandler) checkReview(ctx context.Context) error {
	info, err := reviewmwcli.GetReview(ctx, *h.EntID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("invalid review")
	}
	if *h.ID != info.ID {
		return fmt.Errorf("invalid id")
	}
	if *h.TargetAppID != info.AppID {
		return fmt.Errorf("appid mismatch")
	}
	if info.State != reviewtypes.ReviewState_Wait {
		return fmt.Errorf("current review state can not be updated")
	}
	if *h.State == reviewtypes.ReviewState_Rejected && h.Message == nil {
		return fmt.Errorf("message is must")
	}
	h.review = info
	return nil
}

func (h *updateHandler) getCouponWithdraw(ctx context.Context) error {
	info, err := couponwithdrawmwcli.GetCouponWithdraw(ctx, h.review.ObjectID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("invalid coupon withdraw")
	}
	if info.State != ledgertypes.WithdrawState_Reviewing {
		return fmt.Errorf("withdraw state not reviewing")
	}

	h.CouponWithdrawID = &info.EntID
	h.couponwithdraw = info
	return nil
}

func (h *updateHandler) checkKyc(ctx context.Context) error {
	info, err := kycmwcli.GetKycOnly(ctx, &kycmwpb.Conds{
		AppID:  &basetypes.StringVal{Op: cruder.EQ, Value: h.couponwithdraw.AppID},
		UserID: &basetypes.StringVal{Op: cruder.EQ, Value: h.couponwithdraw.UserID},
	})
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("kyc not approved")
	}
	if info.State != basetypes.KycState_Approved {
		return fmt.Errorf("kyc not approved")
	}

	user, err := usermwcli.GetUser(ctx, h.couponwithdraw.AppID, h.couponwithdraw.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("invalid user")
	}
	return nil
}

func (h *updateHandler) updateReview(ctx context.Context) error {
	if _, err := reviewmwcli.UpdateReview(ctx, &reviewmwpb.ReviewReq{
		ID:         &h.review.ID,
		ReviewerID: h.UserID,
		State:      h.State,
		Message:    h.Message,
	}); err != nil {
		return err
	}
	return nil
}

func (h *updateHandler) checkCoin(ctx context.Context) error {
	coin, err := coinmwcli.GetCoin(ctx, h.couponwithdraw.CoinTypeID)
	if err != nil {
		return err
	}
	if coin == nil {
		return fmt.Errorf("invalid cointypeid")
	}
	return nil
}

func (h *Handler) UpdateCouponWithdrawReview(ctx context.Context) (*npool.CouponWithdrawReview, error) {
	handler := &updateHandler{
		Handler: h,
	}

	if err := handler.checkUser(ctx); err != nil {
		return nil, err
	}
	if err := handler.checkReview(ctx); err != nil {
		return nil, err
	}
	if err := handler.getCouponWithdraw(ctx); err != nil {
		return nil, err
	}
	if err := handler.checkKyc(ctx); err != nil {
		return nil, err
	}
	if err := handler.checkCoin(ctx); err != nil {
		return nil, err
	}
	if err := handler.updateReview(ctx); err != nil {
		return nil, err
	}

	return h.GetCouponWithdrawReview(ctx)
}

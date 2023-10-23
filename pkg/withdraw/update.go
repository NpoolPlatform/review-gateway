package withdraw

import (
	"context"
	"fmt"

	kycmwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/kyc"
	usermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	coinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/coin"
	withdrawmwcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/withdraw"
	"github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	kycmwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/kyc"
	ledgertypes "github.com/NpoolPlatform/message/npool/basetypes/ledger/v1"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	withdrawmwpb "github.com/NpoolPlatform/message/npool/ledger/mw/v2/withdraw"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	reviewmwpb "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

type updateHandler struct {
	*Handler
	review   *reviewmwpb.Review
	withdraw *withdrawmwpb.Withdraw
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
	info, err := reviewmwcli.GetReview(ctx, *h.ReviewID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("invalid review")
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

func (h *updateHandler) getWithdraw(ctx context.Context) error {
	info, err := withdrawmwcli.GetWithdraw(ctx, h.review.ObjectID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("invalid withdraw")
	}
	if info.State != ledgertypes.WithdrawState_Reviewing {
		return fmt.Errorf("withdraw state not reviewing")
	}

	h.WithdrawID = &info.ID
	h.withdraw = info
	return nil
}

func (h *updateHandler) checkKyc(ctx context.Context) error {
	info, err := kycmwcli.GetKycOnly(ctx, &kycmwpb.Conds{
		AppID:  &basetypes.StringVal{Op: cruder.EQ, Value: h.review.AppID},
		UserID: &basetypes.StringVal{Op: cruder.EQ, Value: *h.UserID},
	})
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("kyc review not created")
	}
	if info.ReviewID != *h.ReviewID {
		return fmt.Errorf("reviewid mismatch")
	}
	if info.State != basetypes.KycState_Approved {
		return fmt.Errorf("kyc not approved")
	}

	info1, err := usermwcli.GetUser(ctx, info.AppID, info.UserID)
	if err != nil {
		return err
	}
	if info1 == nil {
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
	coin, err := coinmwcli.GetCoin(ctx, h.withdraw.CoinTypeID)
	if err != nil {
		return err
	}
	if coin == nil {
		return fmt.Errorf("invalid cointypeid")
	}
	if coin.Disabled {
		return fmt.Errorf("coin disabled")
	}
	return nil
}

func (h *Handler) UpdateWithdrawReview(ctx context.Context) (*npool.WithdrawReview, error) {
	handler := &updateHandler{
		Handler: h,
	}

	if err := handler.checkUser(ctx); err != nil {
		return nil, err
	}
	if err := handler.checkReview(ctx); err != nil {
		return nil, err
	}
	if err := handler.getWithdraw(ctx); err != nil {
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

	return h.GetWithdrawReview(ctx)
}

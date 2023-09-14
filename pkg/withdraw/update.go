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
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"
)

//nolint:gocyclo
func (h *Handler) UpdateWithdrawReview(ctx context.Context) (*npool.WithdrawReview, error) {
	reviewID := h.ReviewID.String()
	handler, err := review1.NewHandler(
		ctx,
		review1.WithAppID(h.AppID),
		review1.WithUserID(h.AppID, h.UserID),
		review1.WithReviewID(&reviewID),
		review1.WithState(h.State, h.Message),
		review1.WithMessage(h.Message),
	)
	if err != nil {
		return nil, err
	}

	objID, err := handler.ValidateReview(ctx)
	if err != nil {
		return nil, err
	}

	w, err := withdrawmwcli.GetWithdraw(ctx, objID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, fmt.Errorf("invalid withdraw")
	}

	if w.State != ledgertypes.WithdrawState_Reviewing {
		return nil, fmt.Errorf("not reviewing")
	}

	kyc, err := kycmwcli.GetKycOnly(ctx, &kycmwpb.Conds{
		AppID:  &basetypes.StringVal{Op: cruder.EQ, Value: w.AppID},
		UserID: &basetypes.StringVal{Op: cruder.EQ, Value: w.UserID},
	})
	if err != nil {
		return nil, err
	}
	if kyc == nil {
		return nil, fmt.Errorf("kyc review not created")
	}

	if kyc.State != basetypes.KycState_Approved {
		return nil, fmt.Errorf("kyc review not approved")
	}

	userInfo, err := usermwcli.GetUser(ctx, w.AppID, w.UserID)
	if err != nil {
		return nil, err
	}
	if userInfo == nil {
		return nil, fmt.Errorf("invalid user")
	}

	coin, err := coinmwcli.GetCoin(ctx, w.CoinTypeID)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, fmt.Errorf("invalid cointypeid")
	}
	if coin.Disabled {
		return nil, fmt.Errorf("coin disabled")
	}

	// TODO: make sure review state and withdraw state integrity
	switch *h.State {
	case reviewtypes.ReviewState_Rejected:
	case reviewtypes.ReviewState_Approved:
	default:
		return nil, fmt.Errorf("unknown state")
	}

	if err != nil {
		return nil, err
	}

	if err := handler.UpdateReview(ctx); err != nil {
		return nil, err
	}

	return h.GetWithdrawReview(ctx)
}

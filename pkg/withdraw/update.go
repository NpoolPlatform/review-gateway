package withdraw

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	kycmwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/kyc"
	kycmwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/kyc"

	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"

	"github.com/shopspring/decimal"

	"github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mw/v2"

	withdrawmgrcli "github.com/NpoolPlatform/ledger-manager/pkg/client/withdraw"
	withdrawmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/withdraw"

	ledgermwcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/ledger"
	ledgerdetailmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/detail"

	useraccmwcli "github.com/NpoolPlatform/account-middleware/pkg/client/user"
	useraccmwpb "github.com/NpoolPlatform/message/npool/account/mw/v1/user"

	appcoinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/app/coin"
	coinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/coin"
	appcoinmwpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/app/coin"

	pltfaccmwcli "github.com/NpoolPlatform/account-middleware/pkg/client/platform"
	pltfaccmwpb "github.com/NpoolPlatform/message/npool/account/mw/v1/platform"

	accountmgrpb "github.com/NpoolPlatform/message/npool/account/mgr/v1/account"

	txmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/tx"
	txmwpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/tx"

	"github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	commonpb "github.com/NpoolPlatform/message/npool"

	currvalmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/coin/currency"
	currvalmwpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/coin/currency"

	sphinxproxypb "github.com/NpoolPlatform/message/npool/sphinxproxy"
	sphinxproxycli "github.com/NpoolPlatform/sphinx-proxy/pkg/client"

	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"

	txnotifmgrpb "github.com/NpoolPlatform/message/npool/notif/mgr/v1/notif/tx"
	txnotifcli "github.com/NpoolPlatform/notif-middleware/pkg/client/notif/tx"

	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
)

//nolint:gocyclo
func (h *Handler) UpdateWithdrawReview(ctx context.Context) (*withdraw.WithdrawReview, error) {
	reviewID := h.ReviewID.String()
	handler, err := review1.NewHandler(
		ctx,
		review1.WithAppID(h.AppID),
		review1.WithUserID(h.UserID),
		review1.WithReviewID(&reviewID),
		review1.WithState(h.State, nil),
		review1.WithMessage(h.Message),
	)
	if err != nil {
		return nil, err
	}

	objID, err := handler.ValidateReview(ctx)
	if err != nil {
		return nil, err
	}

	w, err := withdrawmgrcli.GetWithdraw(ctx, objID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, fmt.Errorf("invalid withdraw")
	}

	if w.State != withdrawmgrpb.WithdrawState_Reviewing {
		return nil, fmt.Errorf("not reviewing")
	}

	kyc, err := kycmwcli.GetKycOnly(ctx, &kycmwpb.Conds{
		AppID: &basetypes.StringVal{
			Op:    cruder.EQ,
			Value: w.AppID,
		},
		UserID: &basetypes.StringVal{
			Op:    cruder.EQ,
			Value: w.UserID,
		},
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

	userInfo, err := usercli.GetUser(ctx, w.AppID, w.UserID)
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
		return nil, fmt.Errorf("invalid cointypeid")
	}

	// TODO: make sure review state and withdraw state integrity

	switch *h.State {
	case reviewmgrpb.ReviewState_Rejected:
		err = reject(ctx, w)
	case reviewmgrpb.ReviewState_Approved:
		err = approve(ctx, w)
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

func reject(ctx context.Context, withdrawInfo *withdrawmgrpb.Withdraw) error {
	unlocked, err := decimal.NewFromString(withdrawInfo.Amount)
	if err != nil {
		return err
	}

	state := withdrawmgrpb.WithdrawState_Rejected
	// TODO: move to TX

	if err := ledgermwcli.UnlockBalance(
		ctx,
		withdrawInfo.AppID, withdrawInfo.UserID, withdrawInfo.CoinTypeID,
		ledgerdetailmgrpb.IOSubType_Withdrawal,
		unlocked, decimal.NewFromInt(0),
		fmt.Sprintf(
			`{"WithdrawID":"%v","AccountID":"%v"}`,
			withdrawInfo.ID,
			withdrawInfo.AccountID,
		),
	); err != nil {
		return err
	}

	// Update withdraw state
	u := &withdrawmgrpb.WithdrawReq{
		ID:    &withdrawInfo.ID,
		State: &state,
	}
	_, err = withdrawmgrcli.UpdateWithdraw(ctx, u)
	return err
}

// nolint
func approve(ctx context.Context, withdraw *withdrawmgrpb.Withdraw) error {
	wa, err := useraccmwcli.GetAccountOnly(ctx, &useraccmwpb.Conds{
		AppID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.AppID,
		},
		UserID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.UserID,
		},
		CoinTypeID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.CoinTypeID,
		},
		AccountID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.AccountID,
		},
		UsedFor: &commonpb.Int32Val{
			Op:    cruder.EQ,
			Value: int32(accountmgrpb.AccountUsedFor_UserWithdraw),
		},
		Active: &commonpb.BoolVal{
			Op:    cruder.EQ,
			Value: true,
		},
		Blocked: &commonpb.BoolVal{
			Op:    cruder.EQ,
			Value: false,
		},
	})
	if err != nil {
		return err
	}
	if wa == nil {
		return fmt.Errorf("invalid withdraw account")
	}

	coin, err := appcoinmwcli.GetCoinOnly(ctx, &appcoinmwpb.Conds{
		AppID: &basetypes.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.AppID,
		},
		CoinTypeID: &basetypes.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.CoinTypeID,
		},
	})
	if err != nil {
		return err
	}
	if coin == nil {
		return fmt.Errorf("invalid coin")
	}

	hotacc, err := pltfaccmwcli.GetAccountOnly(ctx, &pltfaccmwpb.Conds{
		CoinTypeID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.CoinTypeID,
		},
		UsedFor: &commonpb.Int32Val{
			Op:    cruder.EQ,
			Value: int32(accountmgrpb.AccountUsedFor_UserBenefitHot),
		},
		Backup: &commonpb.BoolVal{
			Op:    cruder.EQ,
			Value: false,
		},
		Active: &commonpb.BoolVal{
			Op:    cruder.EQ,
			Value: true,
		},
		Blocked: &commonpb.BoolVal{
			Op:    cruder.EQ,
			Value: false,
		},
	})
	if err != nil {
		return err
	}
	if hotacc == nil {
		return fmt.Errorf("invalid hot account")
	}

	bal, err := sphinxproxycli.GetBalance(ctx, &sphinxproxypb.GetBalanceRequest{
		Name:    coin.Name,
		Address: hotacc.Address,
	})
	if err != nil {
		return err
	}
	if bal == nil {
		return fmt.Errorf("invalid balance")
	}

	balance := decimal.RequireFromString(bal.BalanceStr)
	amount := decimal.RequireFromString(withdraw.Amount)

	if balance.Cmp(amount) <= 0 {
		return fmt.Errorf("insufficient funds")
	}

	if coin.ID != coin.FeeCoinTypeID {
		feecoin, err := coinmwcli.GetCoin(ctx, coin.FeeCoinTypeID)
		if err != nil {
			return err
		}
		if feecoin == nil {
			return fmt.Errorf("invalid fee coin")
		}

		bal, err := sphinxproxycli.GetBalance(ctx, &sphinxproxypb.GetBalanceRequest{
			Name:    feecoin.Name,
			Address: hotacc.Address,
		})
		if err != nil {
			return err
		}
		if bal == nil {
			return fmt.Errorf("invalid balance")
		}

		feeAmount, err := decimal.NewFromString(feecoin.LowFeeAmount)
		if err != nil {
			return err
		}

		balance := decimal.RequireFromString(bal.BalanceStr)
		if balance.Cmp(feeAmount) <= 0 {
			return fmt.Errorf("insufficient gas")
		}
	}

	feeAmount, err := decimal.NewFromString(coin.WithdrawFeeAmount)
	if err != nil {
		return err
	}

	if coin.WithdrawFeeByStableUSD {
		curr, err := currvalmwcli.GetCurrencyOnly(ctx, &currvalmwpb.Conds{
			CoinTypeID: &basetypes.StringVal{Op: cruder.EQ, Value: withdraw.CoinTypeID},
		})
		if err != nil {
			return err
		}
		if curr == nil {
			return fmt.Errorf("invalid coin currency")
		}

		val, err := decimal.NewFromString(curr.MarketValueLow)
		if err != nil {
			return err
		}
		if val.Cmp(decimal.NewFromInt(0)) <= 0 {
			return fmt.Errorf("invalid coin currency")
		}

		feeAmount = feeAmount.Div(val)
	}

	amountS := amount.String()
	feeAmountS := feeAmount.String()
	txType := basetypes.TxType_TxWithdraw
	txExtra := fmt.Sprintf(
		`{"AppID":"%v","UserID":"%v","Address":"%v","CoinName":"%v","WithdrawID":"%v"}`,
		withdraw.AppID,
		withdraw.UserID,
		wa.Address,
		coin.Name,
		withdraw.ID,
	)

	tx, err := txmwcli.CreateTx(ctx, &txmwpb.TxReq{
		CoinTypeID:    &withdraw.CoinTypeID,
		FromAccountID: &hotacc.AccountID,
		ToAccountID:   &withdraw.AccountID,
		Amount:        &amountS,
		FeeAmount:     &feeAmountS,
		Extra:         &txExtra,
		Type:          &txType,
	})
	if err != nil {
		return err
	}

	state1 := withdrawmgrpb.WithdrawState_Transferring
	if _, err := withdrawmgrcli.UpdateWithdraw(ctx, &withdrawmgrpb.WithdrawReq{
		ID:                    &withdraw.ID,
		PlatformTransactionID: &tx.ID,
		State:                 &state1,
	}); err != nil {
		return err
	}

	txNotifState := txnotifmgrpb.TxState_WaitSuccess
	txNotifType := basetypes.TxType_TxWithdraw
	_, err = txnotifcli.CreateTx(ctx, &txnotifmgrpb.TxReq{
		TxID:       &tx.ID,
		NotifState: &txNotifState,
		TxType:     &txNotifType,
	})
	if err != nil {
		logger.Sugar().Errorw("CreateTx", "error", err.Error())
	}

	return nil
}

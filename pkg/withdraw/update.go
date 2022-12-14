package withdraw

import (
	"context"
	"fmt"

	kyccli "github.com/NpoolPlatform/appuser-manager/pkg/client/kyc"
	kycpb "github.com/NpoolPlatform/message/npool/appuser/mgr/v2/kyc"

	"github.com/shopspring/decimal"

	"github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	withdrawmgrcli "github.com/NpoolPlatform/ledger-manager/pkg/client/withdraw"
	withdrawmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/withdraw"

	ledgermwcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/ledger"
	ledgerdetailmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/detail"

	useraccmwcli "github.com/NpoolPlatform/account-middleware/pkg/client/user"
	useraccmwpb "github.com/NpoolPlatform/message/npool/account/mw/v1/user"

	appcoinmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/appcoin"
	appcoinmwpb "github.com/NpoolPlatform/message/npool/chain/mw/v1/appcoin"

	pltfaccmwcli "github.com/NpoolPlatform/account-middleware/pkg/client/platform"
	pltfaccmwpb "github.com/NpoolPlatform/message/npool/account/mw/v1/platform"

	accountmgrpb "github.com/NpoolPlatform/message/npool/account/mgr/v1/account"

	txmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/tx"
	txmgrpb "github.com/NpoolPlatform/message/npool/chain/mgr/v1/tx"

	"github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	commonpb "github.com/NpoolPlatform/message/npool"

	currvalmwcli "github.com/NpoolPlatform/chain-middleware/pkg/client/coin/currency"

	sphinxproxypb "github.com/NpoolPlatform/message/npool/sphinxproxy"
	sphinxproxycli "github.com/NpoolPlatform/sphinx-proxy/pkg/client"

	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"
)

func UpdateWithdrawReview(
	ctx context.Context,
	id, appID, reviewerAppID, reviewerID string,
	state reviewmgrpb.ReviewState,
	message string,
) (
	*withdraw.WithdrawReview, error,
) {
	objID, err := review1.ValidateReview(ctx, id, appID, reviewerAppID, reviewerID, state)
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

	kyc, err := kyccli.GetKycOnly(ctx, &kycpb.Conds{
		AppID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: w.AppID,
		},
		UserID: &commonpb.StringVal{
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

	if kyc.State != kycpb.KycState_Approved {
		return nil, fmt.Errorf("kyc review not approved")
	}

	// TODO: make sure review state and withdraw state integrity

	switch state {
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

	if err := review1.UpdateReview(ctx, id, appID, reviewerAppID, reviewerID, state, message); err != nil {
		return nil, err
	}

	return GetWithdrawReview(ctx, id)
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
		AppID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: withdraw.AppID,
		},
		CoinTypeID: &commonpb.StringVal{
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

	feeAmount, err := decimal.NewFromString(coin.WithdrawFeeAmount)
	if err != nil {
		return err
	}

	if coin.WithdrawFeeByStableUSD {
		curr, err := currvalmwcli.GetCoinCurrency(ctx, withdraw.CoinTypeID)
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
	txType := txmgrpb.TxType_TxWithdraw
	txExtra := fmt.Sprintf(
		`{"AppID":"%v","UserID":"%v","Address":"%v","CoinName":"%v","WithdrawID":"%v"}`,
		withdraw.AppID,
		withdraw.UserID,
		wa.Address,
		coin.Name,
		withdraw.ID,
	)

	tx, err := txmwcli.CreateTx(ctx, &txmgrpb.TxReq{
		CoinTypeID:    &withdraw.CoinTypeID,
		FromAccountID: &hotacc.ID,
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

	return nil
}

package withdraw

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/withdraw"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	withdrawmgrcli "github.com/NpoolPlatform/ledger-manager/pkg/client/withdraw"
	withdrawmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/withdraw"

	ledgermwcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/ledger"
	ledgerdetailmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/detail"

	billingcli "github.com/NpoolPlatform/cloud-hashing-billing/pkg/client"
	billingpb "github.com/NpoolPlatform/message/npool/cloud-hashing-billing"

	coininfocli "github.com/NpoolPlatform/sphinx-coininfo/pkg/client"

	sphinxproxypb "github.com/NpoolPlatform/message/npool/sphinxproxy"
	sphinxproxycli "github.com/NpoolPlatform/sphinx-proxy/pkg/client"

	currency "github.com/NpoolPlatform/oracle-manager/pkg/middleware/currency"

	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"

	"github.com/google/uuid"
)

func UpdateWithdrawReview(
	ctx context.Context,
	id, appID, reviewerAppID, reviewerID string,
	state reviewmgrpb.ReviewState,
	message string,
) (
	*npool.WithdrawReview, error,
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

	// TODO: get user kyc state
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

func reject(ctx context.Context, withdraw *withdrawmgrpb.Withdraw) error {
	unlocked, err := decimal.NewFromString(withdraw.Amount)
	if err != nil {
		return err
	}

	state := withdrawmgrpb.WithdrawState_Rejected
	// TODO: move to TX

	if err := ledgermwcli.UnlockBalance(
		ctx,
		withdraw.AppID, withdraw.UserID, withdraw.CoinTypeID,
		ledgerdetailmgrpb.IOSubType_Withdrawal,
		unlocked, decimal.NewFromInt(0),
		fmt.Sprintf(
			`{"WithdrawID":"%v","AccountID":"%v"}`,
			withdraw.ID,
			withdraw.AccountID,
		),
	); err != nil {
		return err
	}

	// Update withdraw state
	u := &withdrawmgrpb.WithdrawReq{
		ID:    &withdraw.ID,
		State: &state,
	}
	_, err = withdrawmgrcli.UpdateWithdraw(ctx, u)
	return err
}

// nolint
func approve(ctx context.Context, withdraw *withdrawmgrpb.Withdraw) error {
	// Check account
	account, err := billingcli.GetAccount(ctx, withdraw.AccountID)
	if err != nil {
		return err
	}
	if account == nil {
		return fmt.Errorf("invalid account")
	}

	// Check account is belong to user and used for withdraw
	wa, err := billingcli.GetWithdrawAccount(ctx, withdraw.AccountID)
	if err != nil {
		return err
	}
	if wa == nil {
		return fmt.Errorf("invalid withdraw account")
	}
	if wa.AppID != withdraw.AppID || wa.UserID != withdraw.UserID {
		return fmt.Errorf("invalid user withdraw account")
	}

	// Check hot wallet balance
	coin, err := coininfocli.GetCoinInfo(ctx, withdraw.CoinTypeID)
	if err != nil {
		return err
	}
	if coin == nil {
		return fmt.Errorf("invalid coin")
	}

	cs, err := billingcli.GetCoinSetting(ctx, coin.ID)
	if err != nil {
		return err
	}
	if cs == nil {
		return fmt.Errorf("invalid coin setting")
	}

	hotacc, err := billingcli.GetAccount(ctx, cs.UserOnlineAccountID)
	if err != nil {
		return err
	}
	if hotacc == nil {
		return fmt.Errorf("invalid account")
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

	price, err := currency.USDPrice(ctx, coin.Name)
	if err != nil {
		return err
	}
	if price <= 0 {
		return fmt.Errorf("invalid coin price")
	}

	const feeUSDAmount = 2
	feeAmount := feeUSDAmount / price

	tx, err := billingcli.CreateTransaction(ctx, &billingpb.CoinAccountTransaction{
		AppID:          withdraw.AppID,
		UserID:         withdraw.UserID,
		CoinTypeID:     withdraw.CoinTypeID,
		GoodID:         uuid.UUID{}.String(),
		FromAddressID:  hotacc.ID,
		ToAddressID:    account.ID,
		Amount:         amount.InexactFloat64(),
		TransactionFee: feeAmount,
		Message:        fmt.Sprintf("user withdraw at %v", time.Now()),
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

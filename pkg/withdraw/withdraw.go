package withdraw

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	withdrawmgrcli "github.com/NpoolPlatform/ledger-manager/pkg/client/withdraw"
	withdrawmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/withdraw"

	ledgermwcli "github.com/NpoolPlatform/ledger-middleware/pkg/client/ledger"
	ledgerdetailmgrpb "github.com/NpoolPlatform/message/npool/ledger/mgr/v1/ledger/detail"

	billingcli "github.com/NpoolPlatform/cloud-hashing-billing/pkg/client"
	billingpb "github.com/NpoolPlatform/message/npool/cloud-hashing-billing"

	coininfocli "github.com/NpoolPlatform/sphinx-coininfo/pkg/client"

	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"

	sphinxproxypb "github.com/NpoolPlatform/message/npool/sphinxproxy"
	sphinxproxycli "github.com/NpoolPlatform/sphinx-proxy/pkg/client"

	currency "github.com/NpoolPlatform/oracle-manager/pkg/middleware/currency"

	"github.com/google/uuid"
)

func UpdateReview(ctx context.Context, id string, state reviewmgrpb.ReviewState) (*npool.Review, error) {
	w, err := withdrawmgrcli.GetWithdraw(ctx, id)
	if err != nil {
		return nil, err
	}

	invalidID := uuid.UUID{}.String()
	if w.PlatformTransactionID != "" && w.PlatformTransactionID != invalidID {
		return nil, fmt.Errorf("transaction ongoing")
	}
	if w.State != withdrawmgrpb.WithdrawState_Reviewing {
		return nil, fmt.Errorf("not reviewing")
	}

	var r *npool.Review

	switch state {
	case reviewmgrpb.ReviewState_Rejected:
		r, err = reject(ctx, w)
	case reviewmgrpb.ReviewState_Approved:
		r, err = approve(ctx, w)
	default:
		return nil, fmt.Errorf("unknown state")
	}

	if err != nil {
		return nil, err
	}

	return post(ctx, w, r)
}

func post(ctx context.Context, withdraw *withdrawmgrpb.Withdraw, review *npool.Review) (*npool.Review, error) {
	user, err := usercli.GetUser(ctx, withdraw.AppID, withdraw.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid user")
	}

	review.EmailAddress = user.EmailAddress
	review.PhoneNO = user.PhoneNO

	return review, nil
}

func reject(ctx context.Context, withdraw *withdrawmgrpb.Withdraw) (*npool.Review, error) {
	unlocked, err := decimal.NewFromString(withdraw.Amount)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Update withdraw state
	u := &withdrawmgrpb.WithdrawReq{
		ID:    &withdraw.ID,
		State: &state,
	}
	_, err = withdrawmgrcli.UpdateWithdraw(ctx, u)
	return &npool.Review{}, err
}

// nolint
func approve(ctx context.Context, withdraw *withdrawmgrpb.Withdraw) (*npool.Review, error) {
	r := &npool.Review{}

	// Check account
	account, err := billingcli.GetAccount(ctx, withdraw.AccountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("invalid account")
	}

	// Check account is belong to user and used for withdraw
	wa, err := billingcli.GetWithdrawAccount(ctx, withdraw.AccountID)
	if err != nil {
		return nil, err
	}
	if wa == nil {
		return nil, fmt.Errorf("invalid withdraw account")
	}
	if wa.AppID != withdraw.AppID || wa.UserID != withdraw.UserID {
		return nil, fmt.Errorf("invalid user withdraw account")
	}

	// Check hot wallet balance
	coin, err := coininfocli.GetCoinInfo(ctx, withdraw.CoinTypeID)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, fmt.Errorf("invalid coin")
	}

	cs, err := billingcli.GetCoinSetting(ctx, coin.ID)
	if err != nil {
		return nil, err
	}
	if cs == nil {
		return nil, fmt.Errorf("invalid coin setting")
	}

	hotacc, err := billingcli.GetAccount(ctx, cs.UserOnlineAccountID)
	if err != nil {
		return nil, err
	}
	if hotacc == nil {
		return nil, fmt.Errorf("invalid account")
	}

	bal, err := sphinxproxycli.GetBalance(ctx, &sphinxproxypb.GetBalanceRequest{
		Name:    coin.Name,
		Address: hotacc.Address,
	})
	if err != nil {
		return nil, err
	}
	if bal == nil {
		return nil, fmt.Errorf("invalid balance")
	}

	balance := decimal.RequireFromString(bal.BalanceStr)
	amount := decimal.RequireFromString(withdraw.Amount)

	if balance.Cmp(amount) <= 0 {
		return nil, fmt.Errorf("insufficient funds")
	}

	price, err := currency.USDPrice(ctx, coin.Name)
	if err != nil {
		return nil, err
	}
	if price <= 0 {
		return nil, fmt.Errorf("invalid coin price")
	}

	const feeUSDAmount = 2
	feeAmount := feeUSDAmount / price

	r.ObjectInfo = fmt.Sprintf(
		`{"ID":"%v","To":"%v","Amount":"%v","CoinName":"%v","FeeAmount":"%v"}`,
		withdraw.ID,
		account.Address,
		withdraw.Amount,
		coin.Name,
		feeAmount,
	)

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
		return nil, err
	}

	state1 := withdrawmgrpb.WithdrawState_Transferring
	if _, err := withdrawmgrcli.UpdateWithdraw(ctx, &withdrawmgrpb.WithdrawReq{
		ID:                    &withdraw.ID,
		PlatformTransactionID: &tx.ID,
		State:                 &state1,
	}); err != nil {
		return nil, err
	}

	return r, nil
}

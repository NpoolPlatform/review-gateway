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

	billingcli "github.com/NpoolPlatform/cloud-hashing-billing/pkg/client"
	billingpb "github.com/NpoolPlatform/message/npool/cloud-hashing-billing"

	coininfocli "github.com/NpoolPlatform/sphinx-coininfo/pkg/client"

	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"

	sphinxproxypb "github.com/NpoolPlatform/message/npool/sphinxproxy"
	sphinxproxycli "github.com/NpoolPlatform/sphinx-proxy/pkg/client"

	currency "github.com/NpoolPlatform/oracle-manager/pkg/middleware/currency"

	"github.com/google/uuid"
)

func UpdateReview(ctx context.Context, id string, state reviewmgrpb.ReviewState) (*npool.Review, error) { //nolint
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

	r := &npool.Review{}

	switch state {
	case reviewmgrpb.ReviewState_Rejected:
		return r, nil
	case reviewmgrpb.ReviewState_Approved:
	default:
		return nil, fmt.Errorf("unknown state")
	}

	// Check account
	account, err := billingcli.GetAccount(ctx, w.AccountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("invalid account")
	}

	// Check account is belong to user and used for withdraw
	was, err := billingcli.GetWithdrawAccounts(ctx, w.AppID, w.UserID)
	if err != nil {
		return nil, err
	}
	found := false
	for _, wa := range was {
		if wa.AccountID == account.ID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("not user's withdraw address")
	}

	// Check hot wallet balance
	coin, err := coininfocli.GetCoinInfo(ctx, w.CoinTypeID)
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
	amount := decimal.RequireFromString(w.Amount)

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
		w.ID,
		account.Address,
		w.Amount,
		coin.Name,
		feeAmount,
	)

	user, err := usercli.GetUser(ctx, w.AppID, w.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid user")
	}

	r.EmailAddress = user.EmailAddress
	r.PhoneNO = user.PhoneNO

	tx, err := billingcli.CreateTransaction(ctx, &billingpb.CoinAccountTransaction{
		AppID:          w.AppID,
		UserID:         w.UserID,
		CoinTypeID:     w.CoinTypeID,
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
		ID:                    &w.ID,
		PlatformTransactionID: &tx.ID,
		State:                 &state1,
	}); err != nil {
		return nil, err
	}

	return r, nil
}

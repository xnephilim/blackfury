package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/ingenuity-build/quicksilver/x/interchainstaking/types"
)

// helper function for ICA tests.
func NewICA() *types.ICAAccount {
	addr := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	accAddr := sdk.AccAddress(addr)
	ica, err := types.NewICAAccount(accAddr.String(), "mercury-1.deposit", "uqck")
	if err != nil {
		panic("failed to create ICA")
	}
	return ica
}

// test that the balance can be set to a valid coin (good denom + non negative value)
func TestAccountSetBalanceGood(t *testing.T) {
	ica := NewICA()
	err := ica.SetBalance(sdk.NewCoins(sdk.NewCoin("uqck", sdk.NewInt(300))))
	require.NoError(t, err, "setbalance failed")
	require.True(t, ica.Balance.AmountOf("uqck").Equal(sdk.NewInt(300)))
}

// test that the balance panics when set to an invalid denomination.
func TestAccountSetBalanceBadDenom(t *testing.T) {
	ica := NewICA()
	require.PanicsWithError(t, "invalid denom: _fail", func() { ica.SetBalance(sdk.NewCoins(sdk.NewCoin("_fail", sdk.NewInt(300)))) })
}

// test that the balance panics when set to a negative number.
func TestAccountSetBalanceNegativeAmount(t *testing.T) {
	ica := NewICA()
	require.PanicsWithError(t, "negative coin amount: -300", func() { ica.SetBalance(sdk.NewCoins(sdk.NewCoin("uqck", sdk.NewInt(-300)))) })
}

// test balance waitground increments and decrements as expected and errors on negative wg value.
func TestIncrementDecrementWg(t *testing.T) {
	ica := NewICA()
	oldWg := ica.BalanceWaitgroup
	ica.IncrementBalanceWaitgroup()
	firstWg := ica.BalanceWaitgroup
	require.Equal(t, oldWg+1, firstWg)
	require.NoError(t, ica.DecrementBalanceWaitgroup())
	secondWg := ica.BalanceWaitgroup
	require.Equal(t, firstWg-1, secondWg)
	require.Error(t, ica.DecrementBalanceWaitgroup())
}
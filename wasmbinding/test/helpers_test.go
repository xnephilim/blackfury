package wasmbinding

import (
	"testing"
	"time"

	"github.com/ingenuity-build/blackfury/app"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CreateTestInput(t *testing.T) (*app.Blackfury, sdk.Context) {
	t.Helper()

	blackfuryApp := app.Setup(t, false)
	ctx := blackfuryApp.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "blackfury-1", Time: time.Now().UTC()})
	return blackfuryApp, ctx
}

// we need to make this deterministic (same every test run), as content might affect gas costs.
func keyPubAddr() (key crypto.PrivKey, pub crypto.PubKey, addr sdk.AccAddress) { //nolint:unparam
	key = ed25519.GenPrivKey()
	pub = key.PubKey()
	addr = sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func RandomAccountAddress() sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func RandomBech32AccountAddress() string {
	return RandomAccountAddress().String()
}

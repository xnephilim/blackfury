package interchaintest

import (
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v5"
	"github.com/strangelove-ventures/interchaintest/v5/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v5/ibc"
	"github.com/strangelove-ventures/interchaintest/v5/testreporter"
	"github.com/strangelove-ventures/interchaintest/v5/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestInterchainStaking TODO
func TestInterchainStaking(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Create chain factory with Blackfury and Juno
	numVals := 3
	numFullNodes := 3

	config, err := createConfig()
	require.NoError(t, err)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "blackfury",
			ChainConfig:   config,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "juno",
			Version:       "v14.1.0",
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
			//ChainConfig: ibc.ChainConfig{
			//	GasPrices: "0.0uatom",
			//},
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	blackfury, juno := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Create relayer factory to utilize the go-relayer
	client, network := interchaintest.DockerSetup(t)

	r := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).Build(t, client, network)

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain().
		AddChain(blackfury).
		AddChain(juno).
		AddRelayer(r, "rly").
		AddLink(interchaintest.InterchainLink{
			Chain1:  blackfury,
			Chain2:  juno,
			Relayer: r,
			Path:    pathBlackfuryJuno,
		})

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()

	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,

		// This can be used to write to the block database which will index all block data e.g. txs, msgs, events, etc.
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Start the relayer
	require.NoError(t, r.StartRelayer(ctx, eRep, pathBlackfuryJuno))
	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				panic(fmt.Errorf("an error occurred while stopping the relayer: %s", err))
			}
		},
	)

	// Create some user accounts on both chains
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), genesisWalletAmount, blackfury, juno)

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, blackfury, juno)
	require.NoError(t, err)

	// Get our Bech32 encoded user addresses
	blackUser, junoUser := users[0], users[1]

	blackUserAddr := blackUser.FormattedAddress()
	junoUserAddr := junoUser.FormattedAddress()
	_ = blackUserAddr
	_ = junoUserAddr

	runSidecars(t, ctx, blackfury, juno)
}

func runSidecars(t *testing.T, ctx context.Context, blackfury, juno *cosmos.CosmosChain) {
	t.Helper()

	runICQ(t, ctx, blackfury, juno)
	// runXCC(t, ctx, blackfury, juno)
}

func runICQ(t *testing.T, ctx context.Context, blackfury, juno *cosmos.CosmosChain) {
	t.Helper()

	var icq *cosmos.SidecarProcess
	for _, sidecar := range blackfury.Sidecars {
		if sidecar.ProcessName == "icq" {
			icq = sidecar
		}
	}
	require.NotNil(t, icq)

	containerCfg := "config.yaml"

	file := fmt.Sprintf(`default_chain: '%s'
chains:
  '%s':
    key: default
    chain-id: '%s'
    rpc-addr: '%s'
    grpc-addr: '%s'
    account-prefix: black
    keyring-backend: test
    gas-adjustment: 1.2
    gas-prices: 0.01ufury
    min-gas-amount: 0
    key-directory: %s/.icq/keys
    debug: false
    timeout: 20s
    block-timeout: 10s
    output-format: json
    sign-mode: direct
  '%s':
    key: default
    chain-id: '%s'
    rpc-addr: '%s'
    grpc-addr: '%s'
    account-prefix: osmo
    keyring-backend: test
    gas-adjustment: 1.2
    gas-prices: 0.01uosmo
    min-gas-amount: 0
    key-directory: %s/.icq/keys
    debug: false
    timeout: 20s
    block-timeout: 10s
    output-format: json
    sign-mode: direct
`,
		blackfury.Config().ChainID,
		blackfury.Config().ChainID,
		blackfury.Config().ChainID,
		blackfury.GetRPCAddress(),
		blackfury.GetGRPCAddress(),
		icq.HomeDir(),
		juno.Config().ChainID,
		juno.Config().ChainID,
		juno.GetRPCAddress(),
		juno.GetGRPCAddress(),
		icq.HomeDir(),
	)

	err := icq.WriteFile(ctx, []byte(file), containerCfg)
	require.NoError(t, err)
	_, err = icq.ReadFile(ctx, containerCfg)
	require.NoError(t, err)

	err = icq.StartContainer(ctx)
	require.NoError(t, err)

	err = icq.Running(ctx)
	require.NoError(t, err)
}

func runXCC(t *testing.T, ctx context.Context, blackfury, juno *cosmos.CosmosChain) {
	t.Helper()

	var xcc *cosmos.SidecarProcess
	for _, sidecar := range blackfury.Sidecars {
		if sidecar.ProcessName == "xcc" {
			xcc = sidecar
		}
	}
	require.NotNil(t, xcc)

	containerCfg := "config.yaml"

	file := fmt.Sprintf(`source_chain: '%s'
chains:
  black-1: '%s'
  juno-1: '%s'
`,
		blackfury.Config().ChainID,
		blackfury.GetRPCAddress(),
		juno.GetRPCAddress(),
	)

	err := xcc.WriteFile(ctx, []byte(file), containerCfg)
	require.NoError(t, err)
	_, err = xcc.ReadFile(ctx, containerCfg)
	require.NoError(t, err)

	err = xcc.StartContainer(ctx)
	require.NoError(t, err)

	err = xcc.Running(ctx)
	require.NoError(t, err)
}

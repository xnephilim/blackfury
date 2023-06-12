package main_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/ingenuity-build/blackfury/app"
	blackfuryd "github.com/ingenuity-build/blackfury/cmd/blackfuryd"
)

func TestInitCmd(t *testing.T) {
	rootCmd, _ := blackfuryd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",             // Test the init cmd
		"blackfury-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
		fmt.Sprintf("--%s=%s", flags.FlagChainID, "blackfury-1"),
	})

	err := svrcmd.Execute(rootCmd, "BLACKFURYD", app.DefaultNodeHome)
	require.NoError(t, err)
}

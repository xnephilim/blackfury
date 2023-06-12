package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/ingenuity-build/blackfury/app"
	cmdcfg "github.com/ingenuity-build/blackfury/cmd/config"
)

func main() {
	cmdcfg.SetupConfig()
	cmdcfg.RegisterDenoms()

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	app.DefaultNodeHome = filepath.Join(userHomeDir, ".blackfuryd")

	rootCmd, _ := NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "BLACKFURYD", app.DefaultNodeHome); err != nil {
		var exitError *server.ErrorCode
		if errors.As(err, &exitError) {
			os.Exit(exitError.Code)
		}

		os.Exit(1)
	}

	os.Exit(0)
}

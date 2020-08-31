package main

import (
	"encoding/json"
	"io"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	genutilcli "github.com/mellaught/cosmos-fun/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
	app "github.com/mellaught/cosmos-fun/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "eond",
		Short:             "Onlife Enterprise Network Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}
	// CLI commands to initialize the chain
	rootCmd.AddCommand(genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.CollectGenTxsCmd(ctx, cdc, auth.GenesisAccountIterator{}, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.MigrateGenesisCmd(ctx, cdc))
	rootCmd.AddCommand(
		genutilcli.GenTxCmd(
			ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{},
			auth.GenesisAccountIterator{}, app.DefaultNodeHome, app.DefaultCLIHome,
		),
	)
	rootCmd.AddCommand(genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics))

	// AddGenesisAccountCmd allows users to add accounts to the genesis file
	rootCmd.AddCommand(AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))

	// AddGenesisAccountCmd allows users to add accounts to the genesis file
	// rootCmd.AddCommand(AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(flags.NewCompletionCmd(rootCmd, true))
	rootCmd.AddCommand(testnetCmd(ctx, cdc, app.ModuleBasics, auth.GenesisAccountIterator{}))
	rootCmd.AddCommand(debug.Cmd(cdc))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "EON", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewEONApp(logger, db, baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)))
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		eonApp := app.NewEONApp(logger, db)
		err := eonApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return eonApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	eonApp := app.NewEONApp(logger, db)

	return eonApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

// // AddGenesisAccountCmd allows users to add accounts to the genesis file
// func AddGenesisAccountCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "add-genesis-account [address] [coins[,coins]]",
// 		Short: "Adds an account to the genesis file",
// 		Args:  cobra.ExactArgs(2),
// 		Long: strings.TrimSpace(`
//   Adds accounts to the genesis file so that you can start a chain with coins in the CLI:flagClientHome
//   $ nsd add-genesis-account cosmos1tse7r2fadvlrrgau3pa0ss7cqh55wrv6y9alwh 1000STAKE,1000mycoin
//   `),
// 		RunE: func(_ *cobra.Command, args []string) error {
// 			addr, err := sdk.AccAddressFromBech32(args[0])
// 			if err != nil {
// 				return err
// 			}
// 			coins, err := sdk.ParseCoins(args[1])
// 			if err != nil {
// 				return err
// 			}
// 			coins.Sort()

// 			var genDoc tmtypes.GenesisDoc
// 			config := ctx.Config
// 			genFile := config.GenesisFile()
// 			if !common.FileExists(genFile) {
// 				return fmt.Errorf("%s does not exist, run `gaiad init` first", genFile)
// 			}
// 			genContents, err := ioutil.ReadFile(genFile)
// 			if err != nil {
// 			}

// 			if err = cdc.UnmarshalJSON(genContents, &genDoc); err != nil {
// 				return err
// 			}

// 			var appState app.GenesisState
// 			if err = cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
// 				return err
// 			}

// 			for _, stateAcc := range appState.Accounts {
// 				if stateAcc.Address.Equals(addr) {
// 					return fmt.Errorf("the application state already contains account %v", addr)
// 				}
// 			}

// 			acc := auth.NewBaseAccountWithAddress(addr)
// 			acc.Coins = coins
// 			appState.Accounts = append(appState.Accounts, &acc)
// 			appStateJSON, err := cdc.MarshalJSON(appState)
// 			if err != nil {
// 				return err
// 			}

// 			return gaiaInit.ExportGenesisFile(genFile, genDoc.ChainID, genDoc.Validators, appStateJSON)
// 		},
// 	}
// 	return cmd
// }

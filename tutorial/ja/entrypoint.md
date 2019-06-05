# エントリポイント

golangでは、規約はプロジェクトの `./cmd`フォルダにバイナリにコンパイルされるファイルを置くことです。あなたのアプリケーションにはあなたが作りたい2つのバイナリがあります。

 -  `nsd`：このバイナリは、p2p接続を維持し、トランザクションを伝播し、ローカルストレージを処理し、ネットワークと対話するためのRPCインタフェースを提供するという点で、` bitcoind`や他の暗号通貨デーモンと似ています。この場合、Tendermintはネットワーキングとトランザクションの順序付けに使用されます。
 -  `nscli`：このバイナリはユーザがあなたのアプリケーションと対話することを可能にするコマンドを提供します。

はじめに、これらのバイナリをインスタンス化する2つのファイルをプロジェクトディレクトリに作成します。

 -  `./cmd/nsd/main.go`
 -  `./cmd/nscli/main.go`

## `nsd`

以下のコードを `nsd/main.go`に追加することから始めます。

> _*NOTE*_：あなたのアプリケーションは今書いたコードをインポートする必要があります。ここではインポートパスがこのリポジトリに設定されています( `github.com/cosmos/sdk-application-tutorial`)。自分のリポジトリをフォローしている場合は、それを反映するようにインポートパスを変更する必要があります( `github.com/ {.Username}/{.Project.Repo}`)。

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"

	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	sdk "github.com/cosmos/cosmos-sdk/types"
	app "github.com/cosmos/sdk-application-tutorial"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
)

// DefaultNodeHomeは、アプリケーションデータと設定が保存されるフォルダを設定します
var DefaultNodeHome = os.ExpandEnv("$HOME/.nsd")

const (
	flagOverwrite = "overwrite"
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "nsd",
		Short:             "nameservice App Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(InitCmd(ctx, cdc))
	rootCmd.AddCommand(AddGenesisAccountCmd(ctx, cdc))
	server.AddCommands(ctx, cdc, rootCmd, newApp, appExporter())

	//フラグを準備して追加する
	executor := cli.PrepareBaseCmd(rootCmd, "NS", DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		//＃870で処理する
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewNameServiceApp(logger, db)
}

func appExporter() server.AppExporter {
	return func(logger log.Logger, db dbm.DB, _ io.Writer, _ int64, _ bool, _ []string) (
		json.RawMessage, []tmtypes.GenesisValidator, error) {
		dapp := app.NewNameServiceApp(logger, db)
		return dapp.ExportAppStateAndValidators()
	}
}

// InitCmdはすべてのファイルをtendermintとapplication用に初期化します
func InitCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis config, priv-validator file, and p2p-node file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			chainID := viper.GetString(client.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}

			_, pk, err := gaiaInit.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			var appState json.RawMessage
			genFile := config.GenesisFile()

			if !viper.GetBool(flagOverwrite) && common.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			genesis := app.GenesisState{
				AuthData: auth.DefaultGenesisState(),
				BankData: bank.DefaultGenesisState(),
			}

			appState, err = codec.MarshalJSONIndent(cdc, genesis)
			if err != nil {
				return err
			}

			_, _, validator, err := SimpleAppGenTx(cdc, pk)
			if err != nil {
				return err
			}

			if err = gaiaInit.ExportGenesisFile(genFile, chainID, []tmtypes.GenesisValidator{validator}, appState); err != nil {
				return err
			}

			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			fmt.Printf("Initialized nsd configuration and bootstrapping files in %s...\n", viper.GetString(cli.HomeFlag))
			return nil
		},
	}

	cmd.Flags().String(cli.HomeFlag, DefaultNodeHome, "node's home directory")
	cmd.Flags().String(client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")

	return cmd
}

// AddGenesisAccountCmdを使用すると、ユーザーはGeneisファイルにアカウントを追加できます
func AddGenesisAccountCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address] [coins[,coins]]",
		Short: "Adds an account to the genesis file",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(`
Adds accounts to the genesis file so that you can start a chain with coins in the CLI:

$ nsd add-genesis-account cosmos1tse7r2fadvlrrgau3pa0ss7cqh55wrv6y9alwh 1000STAKE,1000nametoken
`),
		RunE: func(_ *cobra.Command, args []string) error {
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			coins.Sort()

			var genDoc tmtypes.GenesisDoc
			config := ctx.Config
			genFile := config.GenesisFile()
			if !common.FileExists(genFile) {
				return fmt.Errorf("%s does not exist, run `gaiad init` first", genFile)
			}
			genContents, err := ioutil.ReadFile(genFile)
			if err != nil {
			}

			if err = cdc.UnmarshalJSON(genContents, &genDoc); err != nil {
				return err
			}

			var appState app.GenesisState
			if err = cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
				return err
			}

			for _, stateAcc := range appState.Accounts {
				if stateAcc.Address.Equals(addr) {
					return fmt.Errorf("the application state already contains account %v", addr)
				}
			}

			acc := auth.NewBaseAccountWithAddress(addr)
			acc.Coins = coins
			appState.Accounts = append(appState.Accounts, &acc)
			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			return gaiaInit.ExportGenesisFile(genFile, genDoc.ChainID, genDoc.Validators, appStateJSON)
		},
	}
	return cmd
}

// SimpleAppGenTxは、ノードを最初からvaldiatorにする単純なGenTxコマンドを返します
func SimpleAppGenTx(cdc *codec.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	addr, secret, err := server.GenerateCoinKey()
	if err != nil {
		return
	}

	bz, err := cdc.MarshalJSON(struct {
		Addr sdk.AccAddress `json:"addr"`
	}{addr})
	if err != nil {
		return
	}

	appGenTx = json.RawMessage(bz)

	bz, err = cdc.MarshalJSON(map[string]string{"secret": secret})
	if err != nil {
		return
	}

	cliPrint = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}

	return
}
```

上記のコードについての注意：

 - 上記のコードの大部分は、Tendermint、Cosmos-SDK、そしてあなたのNameserviceモジュールからのCLIコマンドを組み合わせたものです。
 -  `InitCmd`は、設定から生成状態を生成することをアプリに許可します。チェーンブートストラップ処理の詳細については、関数呼び出しを調べてください。
 -  `AddGenesisAccountCmd`は、ジェネシスファイルにアカウントを追加するのに便利です。チェーンスタート時にコインで財布を入れることができます。

## `nscli`

`nscli`コマンドをビルドして終了してください。

> _*NOTE*_：あなたのアプリケーションは今書いたコードをインポートする必要があります。ここではインポートパスがこのリポジトリに設定されています( `github.com/cosmos/sdk-application-tutorial`)。自分のリポジトリをフォローしている場合は、それを反映するようにインポートパスを変更する必要があります( `github.com/{.Username}/{.Project.Repo}`)。

```go
package main

import (
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	app "github.com/cosmos/sdk-application-tutorial"
	nsclient "github.com/cosmos/sdk-application-tutorial/x/nameservice/client"
	nsrest "github.com/cosmos/sdk-application-tutorial/x/nameservice/client/rest"
)

const (
	storeAcc = "acc"
	storeNS  = "nameservice"
)

var defaultCLIHome = os.ExpandEnv("$HOME/.nscli")

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	// SDKの設定ファイルを読み込む
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	mc := []sdk.ModuleClients{
		nsclient.NewModuleClient(storeNS, cdc),
	}

	rootCmd := &cobra.Command{
		Use:   "nscli",
		Short: "nameservice Client",
	}

	//永続フラグに--chain-idを追加して必須マークを付けます
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	//ルート作成コマンド
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(defaultCLIHome),
		queryCmd(cdc, mc),
		txCmd(cdc, mc),
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		client.LineBreak,
		keys.Commands(),
		client.LineBreak,
	
	)

	executor := cli.PrepareMainCmd(rootCmd, "NS", defaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func registerRoutes(rs *lcd.RestServer) {
	rs.CliCtx = rs.CliCtx.WithAccountDecoder(rs.Cdc)
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, storeAcc)
	bank.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	nsrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, storeNS)
}

func queryCmd(cdc *amino.Codec, mc []sdk.ModuleClients) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		authcmd.GetAccountCmd(storeAcc, cdc),
	)

	for _, m := range mc {
		queryCmd.AddCommand(m.GetQueryCmd())
	}

	return queryCmd
}

func txCmd(cdc *amino.Codec, mc []sdk.ModuleClients) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		tx.GetBroadcastCommand(cdc),
		client.LineBreak,
	)

	for _, m := range mc {
		txCmd.AddCommand(m.GetTxCmd())
	}

	return txCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
```

注意：

 - コードはTendermint、Cosmos-SDKおよびあなたのNameserviceモジュールからのCLIコマンドを結合します。
 -  [`cobra` CLI documentation](http://github.com/spf13/cobra)は上記のコードを理解するのに役立ちます。
 - ここで先に定義した `ModuleClient`を見ることができます。
 - ルートが `registerRoutes`関数にどのように含まれているかに注意してください

### これであなたのバイナリは[依存関係管理とあなたのアプリのビルド](dep.md)を扱う時間を定義しました！

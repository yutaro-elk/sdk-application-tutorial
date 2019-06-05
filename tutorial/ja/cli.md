# ネームサービスモジュールCLI

Cosmos SDKはCLI対話のために[`cobra`](https://github.com/spf13/cobra)ライブラリを使用します。このライブラリは、各モジュールが独自のコマンドを公開するのを簡単にします。モジュールとユーザーのCLIのやり取りを定義するには、まず以下のファイルを作成します。

 - `./x/nameservice/client/cli/query.go`
 - `./x/nameservice/client/cli/tx.go`
 - `./x/nameservice/client/module_client.go`

##クエリ

`query.go`から始めてください。ここで、各モジュールのQueriersに`cobra.Command`を定義します(`resolve`と`whois`)：

```go
package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/sdk-application-tutorial/x/nameservice"
	"github.com/spf13/cobra"
)

// GetCmdResolveNameは名前に関する情報を問い合わせます
func GetCmdResolveName(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "resolve [name]",
		Short: "resolve name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			name := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/resolve/%s", queryRoute, name), nil)
			if err != nil {
				fmt.Printf("could not resolve name - %s \n", string(name))
				return nil
			}

			var out nameservice.QueryResResolve
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdWhoisはドメインに関する情報を問い合わせます
func GetCmdWhois(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "whois [name]",
		Short: "Query whois info of name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			name := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/whois/%s", queryRoute, name), nil)
			if err != nil {
				fmt.Printf("could not resolve whois - %s \n", string(name))
				return nil
			}

			var out nameservice.Whois
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdNamesはすべての名前のリストを問い合わせます
func GetCmdNames(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "names",
		Short: "names",
		// Args：cobra.ExactArgs（1）、
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/names", queryRoute), nil)
			if err != nil {
				fmt.Printf("could not get query names\n")
				return nil
			}

			var out nameservice.QueryResNames
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
```

上記のコードについての注意：

 -  CLIは新しい`context`を導入しました：[`CLIContext`](https://godoc.org/github.com/cosmos/cosmos-sdk/client/context#CLIContext)。 CLIとの対話に必要なユーザー入力およびアプリケーション構成に関するデータを運びます。
 - `cliCtx.QueryWithData()`関数に必要な`path`はあなたの問い合わせルーターの名前に直接対応します。
   - パスの最初の部分は、SDKアプリケーションで可能なクエリの種類を区別するために使用されます。`custom`は`Queriers`のためのものです。
   -  2番目の部分(`nameservice`)は問い合わせを送る先のモジュールの名前です。
   - 最後に、呼び出されるモジュール内に特定のクエリアがあります。
   - この例では、4番目の部分はクエリです。 queryパラメータは単純な文字列なので、これはうまくいきます。より複雑なクエリ入力を可能にするには、[`.QueryWithData()`]の2番目の引数を使用する必要があります(https://godoc.org/github.com/cosmos/cosmos-sdk/client/context#CLIContext.QueryWithData)`data`を渡すための関数。この例については、[Stakingモジュールのクエリア](https://github.com/cosmos/cosmos-sdk/blob/develop/x/stake/querier/querier.go#L103)を参照してください。

##トランザクション

クエリのやりとりが定義されたので、`tx.go`のトランザクション生成に移ります。

> _*NOTE*_：あなたのアプリケーションは今書いたコードをインポートする必要があります。ここではインポートパスがこのリポジトリに設定されています(`github.com/cosmos/sdk-application-tutorial/x/nameservice`)。自分のリポジトリをフォローしている場合は、それを反映するようにインポートパスを変更する必要があります(`github.com/ {.Username}/{.Project.Repo}/x/nameservice`)。

```go
package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/sdk-application-tutorial/x/nameservice"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

// GetCmdBuyNameはBuyNameトランザクションを送信するためのCLIコマンドです。
func GetCmdBuyName(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "buy-name [name] [amount]",
		Short: "bid for existing name or claim new name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			msg := nameservice.NewMsgBuyName(args[0], coins, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

// GetCmdSetNameはSetNameトランザクションを送信するためのCLIコマンドです
func GetCmdSetName(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-name [name] [value]",
		Short: "set the value associated with a name that you own",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			msg := nameservice.NewMsgSetName(args[0], args[1], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}
```

上記のコードについての注意：

 - ここでは`authcmd`パッケージが使われています。 [使い方の詳細についてはgodocsにあります](https://godoc.org/github.com/cosmos/cosmos-sdk/x/auth/client/cli#GetAccountDecoder)。 CLIによって制御されているアカウントへのアクセスを提供し、署名を容易にします。

## モジュールクライアント

この機能をエクスポートする最後の部分は`ModuleClient`と呼ばれ、`./x/nameservice/client/module_client.go`に入ります。 [モジュールクライアント](https://godoc.org/github.com/cosmos/cosmos-sdk/types#ModuleClients)は、モジュールがクライアント機能をエクスポートするための標準的な方法を提供します。

> _*NOTE*_：あなたのアプリケーションは今書いたコードをインポートする必要があります。ここではインポートパスがこのリポジトリに設定されています(`github.com/cosmos/sdk-application-tutorial/x/nameservice`)。自分のリポジトリをフォローしている場合は、それを反映するようにインポートパスを変更する必要があります(`github.com/ {.Username}/{.Project.Repo}/x/nameservice`)。

```go
package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	nameservicecmd "github.com/cosmos/sdk-application-tutorial/x/nameservice/client/cli"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

// ModuleClientはこのモジュールからすべてのクライアント機能をエクスポートします
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmdはこのモジュールのcliクエリコマンドを返します
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	//サブコマンドの下でネームサービスクエリをグループ化する
	namesvcQueryCmd := &cobra.Command{
		Use:   "nameservice",
		Short: "Querying commands for the nameservice module",
	}

	namesvcQueryCmd.AddCommand(client.GetCommands(
		nameservicecmd.GetCmdResolveName(mc.storeKey, mc.cdc),
		nameservicecmd.GetCmdWhois(mc.storeKey, mc.cdc),
	)...)

	return namesvcQueryCmd
}

// GetTxCmdはこのモジュールのトランザクションコマンドを返します
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	namesvcTxCmd := &cobra.Command{
		Use:   "nameservice",
		Short: "Nameservice transactions subcommands",
	}

	namesvcTxCmd.AddCommand(client.PostCommands(
		nameservicecmd.GetCmdBuyName(mc.cdc),
		nameservicecmd.GetCmdSetName(mc.cdc),
	)...)

	return namesvcTxCmd
}
```

上記のコードについての注意：

 - この抽象化により、クライアントは標準の方法でモジュールからクライアント機能をインポートできます。これは、エントリポイントを構築するとき(entrypoint.md)に表示されます。
 - このインタフェースに残りの機能(このチュートリアルの次の部分で説明)を追加するための[未解決の問題](https://github.com/cosmos/cosmos-sdk/issues/2955)もあります。

### これで[RESTクライアントがあなたのモジュールと通信するために使用するルート](rest.md)を定義する準備が整いました。

# モジュールをインポートしてアプリケーションを完成させる

モジュールの準備ができたので、他の2つのモジュールと共に`./app.go`ファイルに組み込むことができます[`auth`](https://godoc.org/github.com/cosmos/cosmos- sdk/x/auth)および[`bank`](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank)：

> _*NOTE*_：あなたのアプリケーションは今書いたコードをインポートする必要があります。ここではインポートパスがこのリポジトリに設定されています(`github.com/cosmos/sdk-application-tutorial/x/nameservice`)。自分のリポジトリをフォローしている場合は、それを反映するようにインポートパスを変更する必要があります(`github.com/ {.Username}/{.Project.Repo}/x/nameservice`)。

```go
package app

import (
	"encoding/json"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/sdk-application-tutorial/x/nameservice"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
)
```

次に、`nameServiceApp`構造体に`Keepers`とストアのキーを追加し、それに応じてコンストラクタを更新する必要があります。

```go

const (
	appName = "nameservice"
)

type nameServiceApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	keyMain          *sdk.KVStoreKey
	keyAccount       *sdk.KVStoreKey
	keyNS            *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	keyParams        *sdk.KVStoreKey
	tkeyParams       *sdk.TransientStoreKey

	accountKeeper       auth.AccountKeeper
	bankKeeper          bank.Keeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	paramsKeeper        params.Keeper
	nsKeeper            nameservice.Keeper
}

// NewNameServiceAppはnameServiceAppのコンストラクタ関数です
func NewNameServiceApp(logger log.Logger, db dbm.DB) *nameServiceApp {

	//最初に、異なるモジュールによって共有されるトップレベルのコーデックを定義します
	cdc := MakeCodec()

	// BaseAppはABCIプロトコルを介してTendermintとのやり取りを処理します
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

	//ここで、必要なストアキーを使ってアプリケーションを初期化します。
	var app = &nameServiceApp{
		BaseApp: bApp,
		cdc:     cdc,

		keyMain:          sdk.NewKVStoreKey("main"),
		keyAccount:       sdk.NewKVStoreKey("acc"),
		keyNS:            sdk.NewKVStoreKey("ns"),
		keyFeeCollection: sdk.NewKVStoreKey("fee_collection"),
		keyParams:        sdk.NewKVStoreKey("params"),
		tkeyParams:       sdk.NewTransientStoreKey("transient_params"),
	}

	return app
}
```

現時点では、コンストラクタにはまだ重要なロジックがありません。つまり、次のことが必要です。

 - 必要な各モジュールから必要な`Keepers`をインスタンス化します。
 - それぞれの`Keeper`が必要とする`storeKeys`を生成します。
 - 各モジュールから`Handler`を登録します。このためには`baseapp`の`router`の`AddRoute()`メソッドを使います。
 - 各モジュールからQuerierを登録します。このためには`baseapp`の`queryRouter`の`AddRoute()`メソッドを使います。
 - `baseApp`マルチストアで提供されたキーに`KVStore`をマウントします。
 - 初期アプリケーション状態を定義するための`initChainer`を設定します。

完成したコンストラクタは次のようになります。

```go
// NewNameServiceAppはnameServiceAppのコンストラクタ関数です
func NewNameServiceApp(logger log.Logger, db dbm.DB) *nameServiceApp {

	//最初に、異なるモジュールによって共有されるトップレベルのコーデックを定義します
	cdc := MakeCodec()

	// BaseAppはABCIプロトコルを介してTendermintとのやり取りを処理します
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

	//ここで、必要なストアキーを使ってアプリケーションを初期化します。
	var app = &nameServiceApp{
		BaseApp: bApp,
		cdc:     cdc,

		keyMain:          sdk.NewKVStoreKey("main"),
		keyAccount:       sdk.NewKVStoreKey("acc"),
		keyNS:            sdk.NewKVStoreKey("ns"),
		keyFeeCollection: sdk.NewKVStoreKey("fee_collection"),
		keyParams:        sdk.NewKVStoreKey("params"),
		tkeyParams:       sdk.NewTransientStoreKey("transient_params"),
	}

	// ParamsKeeperはアプリケーションのパラメータ格納を処理します
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams)

	// AccountKeeperがアドレスを処理 - >アカウント検索
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		app.keyAccount,
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	// BankKeeperを使用するとsdk.Coinsインタラクションを実行できます
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)

	// FeeCollectionKeeperは取引手数料を収集し、それらを手数料分配モジュールに提供します。
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(cdc, app.keyFeeCollection)

	// NameserviceKeeperは、このチュートリアルのモジュールのKeeperです。
	//ネームストアとのやり取りを処理します
	app.nsKeeper = nameservice.NewKeeper(
		app.bankKeeper,
		app.keyNS,
		app.cdc,
	)

	// AnteHandlerは署名の検証とトランザクションの前処理を処理します
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper))

	// app.Routerは、各モジュールがルートを登録するメイントランザクションルーターです。
	//銀行とネームサービスのルートをここに登録する
	app.Router().
		AddRoute("bank", bank.NewHandler(app.bankKeeper)).
		AddRoute("nameservice", nameservice.NewHandler(app.nsKeeper))

	// app.QueryRouterは、各モジュールがルートを登録するメインのクエリールーターです。
	app.QueryRouter().
		AddRoute("nameservice", nameservice.NewQuerier(app.nsKeeper)).
		AddRoute("acc", auth.NewQuerier(app.accountKeeper))

	// initChainerはgenesis.jsonファイルをネットワークの初期状態に変換します。
	app.SetInitChainer(app.initChainer)

	app.MountStores(
		app.keyMain,
		app.keyAccount,
		app.keyNS,
		app.keyFeeCollection,
		app.keyParams,
		app.tkeyParams,
	)

	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}
```

> _*NOTE*_：上記のTransientStoreは、永続化されていない状態のKVStoreのメモリ内実装です。

`initChainer`は最初のチェーンスタート時に`genesis.json`のアカウントがどのようにアプリケーション状態にマッピングされるかを定義します。`ExportAppStateAndValidators`関数はアプリケーションの初期状態をブートストラップするのを助けます。今のところ、これらについてどちらも心配する必要はありません。

コンストラクタは`initChainer`関数を登録しますが、まだ定義されていません。先に進んでそれを作成してください。

```go
// GenesisStateは、チェーンの先頭にあるチェーンの状態を表します。初期状態（口座残高）はここに保存されます。
type GenesisState struct {
	AuthData auth.GenesisState  `json:"auth"`
	BankData bank.GenesisState  `json:"bank"`
	Accounts []*auth.BaseAccount`json:"accounts"`
}

func (app *nameServiceApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(GenesisState)
	err := app.cdc.UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err)
	}

	for _, acc := range genesisState.Accounts {
		acc.AccountNumber = app.accountKeeper.GetNextAccountNumber(ctx)
		app.accountKeeper.SetAccount(ctx, acc)
	}

	auth.InitGenesis(ctx, app.accountKeeper, app.feeCollectionKeeper, genesisState.AuthData)
	bank.InitGenesis(ctx, app.bankKeeper, genesisState.BankData)

	return abci.ResponseInitChain{}
}

// ExportAppStateAndValidatorsが処理を行います
func (app *nameServiceApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
	ctx := app.NewContext(true, abci.Header{})
	accounts := []*auth.BaseAccount{}

	appendAccountsFn := func(acc auth.Account) bool {
		account := &auth.BaseAccount{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}

		accounts = append(accounts, account)
		return false
	}

	app.accountKeeper.IterateAccounts(ctx, appendAccountsFn)

	genState := GenesisState{
		Accounts: accounts,
		AuthData: auth.DefaultGenesisState(),
		BankData: bank.DefaultGenesisState(),
	}

	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}

	return appState, validators, err
}
```

最後に、あなたので使われているすべてのモジュールを正しく登録するアミノ[`*codec.Codec`](https://godoc.org/github.com/cosmos/cosmos-sdk/codec#Codec)を生成するためのヘルパー関数を追加します。応用：

```go
// MakeCodecはAminoに必要なコーデックを生成します
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	nameservice.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
```

### あなたのモジュールを含むアプリケーションを作成したので、[エントリポイントを構築](entrypoint.md)しましょう。

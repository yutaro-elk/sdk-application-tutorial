＃モジュールをインポートしてアプリケーションを完成させる

モジュールの準備ができたので、他の2つのモジュールと共に `./app.go`ファイルに組み込むことができます[` auth`](https://godoc.org/github.com/cosmos/cosmos- sdk/x/auth)および[`bank`](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank)：

> _*NOTE*_：あなたのアプリケーションは今書いたコードをインポートする必要があります。ここではインポートパスがこのリポジトリに設定されています( `github.com/cosmos/sdk-application-tutorial/x/nameservice`)。自分のリポジトリをフォローしている場合は、それを反映するようにインポートパスを変更する必要があります( `github.com/ {.Username}/{.Project.Repo}/x/nameservice`)。

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

次に、 `nameServiceApp`構造体に` Keepers`とストアのキーを追加し、それに応じてコンストラクタを更新する必要があります。

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

// NewNameServiceApp is a constructor function for nameServiceApp
func NewNameServiceApp(logger log.Logger, db dbm.DB) *nameServiceApp {

	// First define the top level codec that will be shared by the different modules
	cdc := MakeCodec()

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

	// Here you initialize your application with the store keys it requires
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

 - 必要な各モジュールから必要な `Keepers`をインスタンス化します。
 - それぞれの `Keeper`が必要とする` storeKeys`を生成します。
 - 各モジュールから `Handler`を登録します。このためには `baseapp`の` router`の `AddRoute()`メソッドを使います。
 - 各モジュールからQuerierを登録します。このためには `baseapp`の` queryRouter`の `AddRoute()`メソッドを使います。
 -  `baseApp`マルチストアで提供されたキーに` KVStore`をマウントします。
 - 初期アプリケーション状態を定義するための `initChainer`を設定します。

完成したコンストラクタは次のようになります。

```go
// NewNameServiceApp is a constructor function for nameServiceApp
func NewNameServiceApp(logger log.Logger, db dbm.DB) *nameServiceApp {

	// First define the top level codec that will be shared by the different modules
	cdc := MakeCodec()

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

	// Here you initialize your application with the store keys it requires
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

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams)

	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		app.keyAccount,
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)

	// The FeeCollectionKeeper collects transaction fees and renders them to the fee distribution module
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(cdc, app.keyFeeCollection)

	// The NameserviceKeeper is the Keeper from the module for this tutorial
	// It handles interactions with the namestore
	app.nsKeeper = nameservice.NewKeeper(
		app.bankKeeper,
		app.keyNS,
		app.cdc,
	)

	// The AnteHandler handles signature verification and transaction pre-processing
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper))

	// The app.Router is the main transaction router where each module registers its routes
	// Register the bank and nameservice routes here
	app.Router().
		AddRoute("bank", bank.NewHandler(app.bankKeeper)).
		AddRoute("nameservice", nameservice.NewHandler(app.nsKeeper))

	// The app.QueryRouter is the main query router where each module registers its routes
	app.QueryRouter().
		AddRoute("nameservice", nameservice.NewQuerier(app.nsKeeper)).
		AddRoute("acc", auth.NewQuerier(app.accountKeeper))

	// The initChainer handles translating the genesis.json file into initial state for the network
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

`initChainer`は最初のチェーンスタート時に` genesis.json`のアカウントがどのようにアプリケーション状態にマッピングされるかを定義します。 `ExportAppStateAndValidators`関数はアプリケーションの初期状態をブートストラップするのを助けます。今のところ、これらについてどちらも心配する必要はありません。

コンストラクタは `initChainer`関数を登録しますが、まだ定義されていません。先に進んでそれを作成してください。

```go
// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState struct {
	AuthData auth.GenesisState   `json:"auth"`
	BankData bank.GenesisState   `json:"bank"`
	Accounts []*auth.BaseAccount `json:"accounts"`
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

// ExportAppStateAndValidators does the things
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

最後に、あなたので使われているすべてのモジュールを正しく登録するアミノ[`* codec.Codec`](https://godoc.org/github.com/cosmos/cosmos-sdk/codec#Codec)を生成するためのヘルパー関数を追加します。応用：

```go
// MakeCodec generates the necessary codecs for Amino
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

###あなたのモジュールを含むアプリケーションを作成したので、[エントリポイントを構築](entrypoint.md)しましょう。

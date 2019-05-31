＃ネームサービスモジュール休息インタフェース

あなたのモジュールは、モジュールの機能へのプログラムによるアクセスを可能にするためにRESTインターフェースを公開することもできます。はじめに、HTTPハンドラを保持するファイルを作成します。

 -  `./x/nameservice/client/rest/rest.go`

始めるために `imports`と` const`を追加してください：

> _*NOTE*_：あなたのアプリケーションは今書いたコードをインポートする必要があります。ここではインポートパスがこのリポジトリに設定されています( `github.com/cosmos/sdk-application-tutorial/x/nameservice`)。自分のリポジトリをフォローしている場合は、それを反映するようにインポートパスを変更する必要があります( `github.com/ {.Username}/{.Project.Repo}/x/nameservice`)。

```go
package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/sdk-application-tutorial/x/nameservice"

	"github.com/gorilla/mux"
)

const (
	restName = "name"
)
```

### RegisterRoutes

まず、 `RegisterRoutes`関数であなたのモジュール用のRESTクライアントインターフェースを定義します。名前空間が他のモジュールのルートと衝突しないように、すべてのルートを自分のモジュール名で始めます。

```go
// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/names", storeName), buyNameHandler(cdc, cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/names", storeName), setNameHandler(cdc, cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/names/{%s}", storeName, restName), resolveNameHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/names/{%s}/whois", storeName, restName), whoIsHandler(cdc, cliCtx, storeName)).Methods("GET")
}
```

###クエリハンドラ

次に、上記のハンドラを定義します。これらは前に定義したCLIメソッドと非常によく似ています。クエリ `whois`と` resolve`から始めましょう：

```go
func resolveNameHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restName]

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/resolve/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func whoIsHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restName]

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/whois/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}


func namesHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/names", storeName), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}
```

上記のコードについての注意：

 - データを取得するために同じ `cliCtx.QueryWithData`関数を使っていることに注意してください
 - これらの機能は、対応するCLI機能とほぼ同じです。

### Txハンドラ

では `buyName`と` setName`トランザクションルートを定義しましょう。これらが実際に購入して名前を設定するトランザクションを送信していないことに注意してください。それはセキュリティ問題であるであろう要求と共にパスワードを送ることを必要とするでしょう。代わりに、これらのエンドポイントはそれぞれの特定のトランザクションを構築して返します。その後、それらは安全な方法で署名され、その後に `/ txs`のような標準のエンドポイントを使ってネットワークにブロードキャストされます。

```go
type buyNameReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Name    string       `json:"name"`
	Amount  string       `json:"amount"`
	Buyer   string       `json:"buyer"`
}

func buyNameHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req buyNameReq

		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(req.Buyer)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		coins, err := sdk.ParseCoins(req.Amount)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := nameservice.NewMsgBuyName(req.Name, coins, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type setNameReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Name    string       `json:"name"`
	Value   string       `json:"value"`
	Owner   string       `json:"owner"`
}

func setNameHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req setNameReq
		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := nameservice.NewMsgSetName(req.Name, req.Value, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
```

上記のコードについての注意：

 -  [`BaseReq`](https://godoc.org/github.com/cosmos/cosmos-sdk/client/utils#BaseReq)には、トランザクションを作成するための基本的な必須フィールド(使用するキー、デコード方法)が含まれています。それは、あなたがどのチェーンにいるのかなど...そして示されているように埋め込まれるように設計されています。
 -  `baseReq.ValidateBasic`と` clientrest.CompleteAndBroadcastTxREST`はレスポンスコードの設定をあなたに代わって処理するので、これらの関数を使うときにエラーや成功を処理することについて心配する必要はありません。

###今すぐあなたのモジュールはそれが[あなたのCosmos SDKアプリケーションに組み込まれる](./app-complete.md)されるために必要な全てを持っています！

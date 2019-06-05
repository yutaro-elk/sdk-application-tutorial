# クエリ

`./x/nameservice/querier.go`ファイルを作成することから始めます。これは、アプリケーション状態のユーザーに対してどのクエリを実行できるかを定義する場所です。あなたの `nameservice`モジュールは二つのクエリを公開します：

 -  `resolve`：これは` name`を取り、 `nameservice`によって格納されている` value`を返します。これはDNSクエリに似ています。
 -  `whois`：これは` name`を取り、その名前の `price`、` value`、そして `owner`を返します。あなたがそれらを購入したいときにいくらの名前がかかるかを把握するために使用されます。

このモジュールへの問い合わせのためのサブルーターとして機能する `NewQuerier`関数を定義することから始めます(` NewHandler`関数に似ています)。クエリ用の `Msg`に似たインターフェースはないので、switchステートメントのケースを手動で定義する必要があります(クエリの` .Route() `関数から引き出すことはできません)。

```go
package nameservice

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// nameserviceクエリによってサポートされるクエリエンドポイント
const (
	QueryResolve = "resolve"
	QueryWhois   = "whois"
	QueryNames   = "names"
)

// NewQuerierは、状態照会用のモジュールレベルルーターです。
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryResolve:
			return queryResolve(ctx, path[1:], req, keeper)
		case QueryWhois:
			return queryWhois(ctx, path[1:], req, keeper)
		case QueryNames:
			return queryNames(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown nameservice query endpoint")
		}
	}
}
```

ルータが定義されたので、各クエリの入力と返り値を定義します。

```go
// nolint: unparam
func queryResolve(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	name := path[0]

	value := keeper.ResolveName(ctx, name)

	if value == "" {
		return []byte{}, sdk.ErrUnknownRequest("could not resolve name")
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, QueryResResolve{value})
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

// 解決クエリの結果
type QueryResResolve struct {
	Value string `json:"value"`
}

// fmt.Stringerを実装
func (r QueryResResolve) String() string {
	return r.Value
}

// nolint: unparam
func queryWhois(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	name := path[0]

	whois := keeper.GetWhois(ctx, name)

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, whois)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

// fmt.Stringerを実装する
func (w Whois) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Owner: %s
Value: %s
Price: %s`, w.Owner, w.Value, w.Price))
}

func queryNames(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var namesList QueryResNames

	iterator := keeper.GetNamesIterator(ctx)

	for ; iterator.Valid(); iterator.Next() {
		name := string(iterator.Key())
		namesList = append(namesList, name)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, namesList)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

// 名前クエリのクエリ結果ペイロード
type QueryResNames []string

// fmt.Stringerを実装する
func (n QueryResNames) String() string {
	return strings.Join(n[:], "\n")
}
```

上記のコードについての注意：

 - ここであなたの `Keeper`のゲッターとセッターが多用されています。このモジュールを使用する他のアプリケーションを構築するときには、戻って必要な状態の部分にアクセスするためにさらにゲッター/セッターを定義する必要があるかもしれません。
 - 慣例により、各出力型はJSON整列化可能かつ文字列化可能(Golangの `fmt.Stringer`インターフェースを実装する)の両方であるべきです。返されるバイトは、出力結果のJSONエンコーディングです。
   - それで、 `resolve`の出力タイプのために、解決文字列をJSON整列化可能で` .String() `メソッドを持つ` QueryResResolve`と呼ばれる構造体にラップします。
   -  Whoisの出力では、通常のWhois構造体はすでにJSONマーシャリング可能ですが、それに `.String()`メソッドを追加する必要があります。
   -  namesクエリの出力と同じですが、 `[]string`は既にネイティブにマーシャリング可能ですが、それに` .String() `メソッドを追加したいと思います。

### モジュールの状態を変化させて表示する方法がありましたので、最後に仕上げましょう。あなたの型を[Amino encoding format next](./codec.md)で登録する!

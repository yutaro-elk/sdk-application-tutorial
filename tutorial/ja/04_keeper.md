# The Keeper

Cosmos SDKモジュールの中心は`Keeper`と呼ばれるものです。ストアとのやり取りを処理し、モジュール間のやり取りのための他のキーパーへの参照を持ち、そしてモジュールのコア機能の大部分を含みます。

## Keeper 構造体

SDKモジュールを起動するには、新しい`./x/nameservice/keeper.go`ファイルに`nameservice.Keeper`を定義してください。

```go
package nameservice

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeperデータストレージへのリンクを維持し、ステートマシンのさまざまな部分のゲッター/セッターメソッドを公開します。
type Keeper struct {
	coinKeeper bank.Keeper

	storeKey  sdk.StoreKey // sdk.Contextからストアにアクセスするための公開されていないキー

	cdc *codec.Codec // バイナリエンコーディング/デコーディング用のワイヤーコーデック。
}
```

上記のコードについての注意点

* 3種類の`cosmos-sdk`パッケージがインポートされています：
	-  [`codec`](https://godoc.org/github.com/cosmos/cosmos-sdk/codec) - `codec`はCosmosエンコーディングフォーマットで動作するためのツールを提供します。[Amino](https://github.com/tendermint/go-amino)
	-  [`bank`](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank) - `bank`モジュールは口座と硬貨振替を制御します。
	-  [`types`](https://godoc.org/github.com/cosmos/cosmos-sdk/types) - `types`にはSDKでよく使われる型が含まれています。
* `Keeper`構造体このキーパーには、いくつかの重要な部分があります。
	-  [`bank.Keeper`](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank#Keeper) - これは`bank`モジュールからの`Keeper`への参照です。それを含めることで、このモジュールのコードが`bank`モジュールから関数を呼び出せるようになります。 SDKは、[オブジェクト機能](https://en.wikipedia.org/wiki/Object-capability_model)アプローチを使用してアプリケーション状態のセクションにアクセスします。これは、開発者が最小権限のアプローチを採用して、不良または悪意のあるモジュールの機能が、アクセスする必要がない状態の部分に影響を与えることを制限することを可能にするためです。
	-  [`*codec.Codec`](https://godoc.org/github.com/cosmos/cosmos-sdk/codec#Codec) - これはAminoがバイナリのエンコードとデコードに使用するコーデックへのポインタです。構造体
	-  [`sdk.StoreKey`](https://godoc.org/github.com/cosmos/cosmos-sdk/types#StoreKey) - これは持続的な`sdk.KVStore`へのアクセスをゲートするストアキーです。あなたのアプリケーションの状態：名前が指し示すWhois構造体(すなわち`map[name]Whois`)。

## ゲッターとセッター

今度は`Keeper`を通してStoreと対話するためのメソッドを追加する時が来ました。まず、与えられた名前が解決するWhoisを設定する関数を追加します。

```go
// Whoisメタデータ構造体全体を名前に設定します
func (k Keeper) SetWhois(ctx sdk.Context, name string, whois Whois) {
	if whois.Owner.Empty() {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(name), k.cdc.MustMarshalBinaryBare(whois))
}
```

このメソッドでは、まず`Keeper`の`storeKey`を使って`map [name] Whois`のstoreオブジェクトを取得します。

> _*NOTE*_：この関数は[`sdk.Context`](https://godoc.org/github.com/cosmos/cosmos-sdk/types#Context)を使います。このオブジェクトは`blockHeight`や`chainID`のような状態のいくつかの重要な部分にアクセスするための関数を保持しています。

次に、`.Set([]byte、[]byte)`メソッドを使って`<name、whois>`ペアをストアに挿入します。ストアは`[]byte`のみを取るので、ストアに挿入される`Whois`構造体を`[]byte`に整列化するためにAminoと呼ばれるCosmos SDKエンコーディングライブラリを使います。

Whoisの所有者フィールドが空の場合、存在するすべての名前に所有者が必要なので、ストアには何も書き込みません。

次に、名前を解決するためのメソッドを追加します(すなわち、`name`の`Whois`を検索します)。

```go
// 名前のWhoisメタデータ構造体全体を取得します
func (k Keeper) GetWhois(ctx sdk.Context, name string) Whois {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(name)) {
		return NewWhois()
	}
	bz := store.Get([]byte(name))
	var whois Whois
	k.cdc.MustUnmarshalBinaryBare(bz, &whois)
	return whois
}
```

ここでは、`SetName`メソッドのように、まず`StoreKey`を使ってストアにアクセスします。次に、storeキーに対して`Set`メソッドを使う代わりに、`.Get([]byte)[]byte`メソッドを使います。関数へのパラメータとして、keyを渡します。それは`[]byte`にキャストされた`name`文字列で、結果を`[]byte`の形で返します。ここでもまたAminoを使用しますが、今回はバイトスライスを「Whois」構造体にアンマーシャルして返します。

現在ストアに名前が存在しない場合は、minimumPriceが初期化されている新しいWhoisを返します。

今回は、名前に基づいてストアから特定のパラメータを取得するための関数を追加しました。しかし、ストアのゲッターとセッターを書き換える代わりに、`GetWhois`と`SetWhois`関数を再利用します。たとえば、フィールドを設定するには、まずWhoisデータ全体を取得し、特定のフィールドを更新してから、新しいバージョンをストアに戻します。
```go
// ResolveName  - 名前が解決される文字列を返します
func (k Keeper) ResolveName(ctx sdk.Context, name string) string {
	return k.GetWhois(ctx, name).Value
}

// SetName  - 名前が解決される値の文字列を設定します
func (k Keeper) SetName(ctx sdk.Context, name string, value string) {
	whois := k.GetWhois(ctx, name)
	whois.Value = value
	k.SetWhois(ctx, name, whois)
}

// HasOwner  - 名前が既に所有者を持っているかどうかを返します
func (k Keeper) HasOwner(ctx sdk.Context, name string) bool {
	return !k.GetWhois(ctx, name).Owner.Empty()
}

// GetOwner  - 名前の現在の所有者を取得します
func (k Keeper) GetOwner(ctx sdk.Context, name string) sdk.AccAddress {
	return k.GetWhois(ctx, name).Owner
}

// SetOwner  - 名前の現在の所有者を設定します
func (k Keeper) SetOwner(ctx sdk.Context, name string, owner sdk.AccAddress) {
	whois := k.GetWhois(ctx, name)
	whois.Owner = owner
	k.SetWhois(ctx, name, whois)
}

// GetPrice  - 名前の現在の価格を取得します。価格がまだ存在しない場合は、1nametokenに設定します。
func (k Keeper) GetPrice(ctx sdk.Context, name string) sdk.Coins {
	return k.GetWhois(ctx, name).Price
}

// SetPrice  - 名前の現在の価格を設定します
func (k Keeper) SetPrice(ctx sdk.Context, name string, price sdk.Coins) {
	whois := k.GetWhois(ctx, name)
	whois.Price = price
	k.SetWhois(ctx, name, whois)
}
```
SDKには`sdk.Iterator`と呼ばれる機能も含まれています。これは、ストア内の特定の場所にあるすべての`<Key、Value>`ペアに対してイテレータを返します。
ストア内に存在するすべての名前のイテレータを取得するための関数を追加します。

```go
// キーが名前で値がwhoisであるすべての名前のイテレータを取得します
func (k Keeper) GetNamesIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte{})
}
```

`./x/nameservice/keeper.go`ファイルに必要な最後のコードは`Keeper`のコンストラクタ関数です。

```go
// NewKeeperはネームサービスKeeperの新しいインスタンスを作成します
func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}
```

### 次に、[`Msgs`と`Handlers`](05_msgs-handlers.md)を使って、ユーザーがあなたの新しいストアとどのようにやり取りするのかを説明します。
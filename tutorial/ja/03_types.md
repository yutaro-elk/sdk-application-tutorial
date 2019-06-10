# 型

最初にやることは、名前のすべてのメタデータを保持する構造体を定義することです。この構造体をICANN DNS用語の後でWhoisと呼びます。

## `types.go`

あなたのモジュールのための関税タイプを保持するためにファイル`./x/nameservice/types.go`を作成することから始めてください。 Cosmos SDKアプリケーションでは、規約はモジュールが`./x/`フォルダにあることです。

## Whois

各名前には3つのデータが関連付けられています。
 - Value - 名前が解決される値。これは単なる任意の文字列ですが、将来的にはIPアドレス、DNSゾーンファイル、ブロックチェーンアドレスなどの特定の形式に合うように変更することができます。
 - Owner - 名前の現在の所有者のアドレス
 - Price - 名前を買うために支払う必要がある価格


SDKモジュールを起動するには、`./x/nameservice/types.go`に`nameservice.Whois`構造体を定義してください。

```go
package nameservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Whoisはnameのすべてのメタデータを含む構造体です。
type Whois struct {
	Value string        `json:"value"`
	Owner sdk.AccAddress`json:"owner"`
	Price sdk.Coins     `json:"price"`
}
```

[Design doc](./01_app-design.md)で説明したように、nameにまだ所有者がいない場合は、MinPriceを使用して名前を初期化します。

```go
// 以前に所有されたことがないnameの初期開始価格
var MinNamePrice = sdk.Coins{sdk.NewInt64Coin("nametoken", 1)}

// 価格としてminpriceを持つ新しいWhoisを返します
func NewWhois() Whois {
	return Whois{
		Price: MinNamePrice,
	}
}
```

### 今度は[Keeper for the module](./04_keeper.md)を書くことに移ります。
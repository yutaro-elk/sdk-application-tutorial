# SetName

## `メッセージ`

SDKの`Msgs`の命名規則は`Msg {.Action}`です。実装する最初のアクションは`SetName`ですので、それを`MsgSetName`と呼びます。この`Msg`は名前の所有者がリゾルバ内でその名前の戻り値を設定することを可能にします。`./x/nameservice/msgs.go`という名前の新しいファイルに`MsgSetName`を定義することから始めます。

```go
package nameservice

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetNameはSetNameメッセージを定義します
type MsgSetName struct {
	Name string
	Value  string
	Owner  sdk.AccAddress
}

// NewMsgSetNameはMsgSetNameのコンストラクタ関数です。
func NewMsgSetName(name string, value string, owner sdk.AccAddress) MsgSetName {
	return MsgSetName{
		Name: name,
		Value:  value,
		Owner:  owner,
	}
}
```

`MsgSetName`は名前の値を設定するのに必要な3つの属性を持ちます。

 - `name` - 設定しようとしている名前。
 - `value` - 名前が解決するもの
 - `owner` - その名前の所有者。

次に`Msg`インターフェースを実装します。

```go
// ルートはモジュールの名前を返すべきです
func (msg MsgSetName) Route() string { return "nameservice" }

// 型はアクションを返すべきです
func (msg MsgSetName) Type() string { return "set_name"}
```

上記の関数はSDKによって`Msgs`を適切なモジュールにルーティングして処理するために使用されます。また、索引付けに使用されるデータベースタグに、判読可能な名前を追加しています。

```go
// ValidateBasicはメッセージに対してステートレスチェックを実行します。
func (msg MsgSetName) ValidateBasic() sdk.Error {
	if msg.Owner.Empty() {
		return sdk.ErrInvalidAddress(msg.Owner.String())
	}
	if len(msg.Name) == 0 || len(msg.Value) == 0 {
		return sdk.ErrUnknownRequest("Name and/or Value cannot be empty")
	}
	return nil
}
```

`ValidateBasic`は`Msg`の妥当性についてのいくつかの基本的な**ステートレス**チェックを提供するために使用されます。この場合は、どの属性も空でないことを確認してください。ここでは`sdk.Error`型の使用に注意してください。 SDKは、アプリケーション開発者が頻繁に遭遇する一連のエラータイプを提供します。

```go
// GetSignBytesは署名のためにメッセージをエンコードします
func (msg MsgSetName) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}
```

`GetSignBytes`は署名のために`Msg`がどのようにエンコードされるかを定義します。ほとんどの場合、これはソートされたJSONへの整列化を意味します。出力は変更しないでください。

```go
// GetSignersは誰の署名が必要かを定義します
func (msg MsgSetName) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
```

`GetSigners`は、それが有効であるために誰の署名が`Tx`に要求されるかを定義します。この場合、例えば、`MsgSetName`は名前が指すものをリセットしようとするときに`Owner`がトランザクションに署名することを要求します。

## `ハンドラ`

`MsgSetName`が指定されたので、次のステップはこのメッセージが受信された時にとるべき行動を定義することです。これが`handler`の役割です。

新しいファイル(`./x/nameservice/handler.go`)では、以下のコードで始めます。

```go
package nameservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandlerは "nameservice"タイプのメッセージ用のハンドラを返します。
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetName:
			return handleMsgSetName(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
```

`NewHandler`は本質的にこのモジュールに入ってくるメッセージを適切なハンドラに送るサブルータです。現時点では、`Msg`/`Handler`は1つだけです。

さて、`handleMsgSetName`で`MsgSetName`メッセージを処理するための実際のロジックを定義する必要があります。

> _*NOTE*_：SDKのハンドラ名の命名規則は`handleMsg {.Action}`です。

```go
// メッセージを処理して名前を設定する
func handleMsgSetName(ctx sdk.Context, keeper Keeper, msg MsgSetName) sdk.Result {
	if !msg.Owner.Equals(keeper.GetOwner(ctx, msg.Name)) { // msg送信者が現在の所有者と同じかどうかを確認します
		return sdk.ErrUnauthorized("Incorrect Owner").Result() // そうでなければ、エラーを投げます
	}
	keeper.SetName(ctx, msg.Name, msg.Value) // その場合は、名前をmsgで指定された値に設定してください。
	return sdk.Result{}                      // return
}
```

この関数では、`Msg`送信者が実際に名前の所有者(`keeper.GetOwner`)であるかどうか確認してください。もしそうなら、彼らは`Keeper`の関数を呼び出すことによって名前を設定することができます。そうでない場合は、エラーをスローしてそれをユーザーに返します。

### すばらしい、今所有者は`SetName`sを持つことができます！しかし、名前にまだ所有者がいない場合はどうなりますか？あなたのモジュールは、ユーザが名前を買うための方法を必要としています！定義しましょう[`BuyName`メッセージを定義します](./buy-name.md).

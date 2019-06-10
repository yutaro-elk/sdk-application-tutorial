# BuyName

## メッセージ

今度は名前を購入するための`Msg`を定義し、それを`./x/nameservice/msgs.go`ファイルに追加します。このコードは`SetName`と非常によく似ています。

```go
// MsgBuyNameはBuyNameメッセージを定義します
type MsgBuyName struct {
	Name string
	Bid    sdk.Coins
	Buyer  sdk.AccAddress
}

// NewMsgBuyNameはMsgBuyNameのコンストラクタ関数です。
func NewMsgBuyName(name string, bid sdk.Coins, buyer sdk.AccAddress) MsgBuyName {
	return MsgBuyName{
		Name: name,
		Bid:    bid,
		Buyer:  buyer,
	}
}

// ルートはモジュールの名前を返すべきです
func (msg MsgBuyName) Route() string { return "nameservice" }

// 型はアクションを返すべきです
func (msg MsgBuyName) Type() string { return "buy_name" }

// ValidateBasicはメッセージに対してステートレスチェックを実行します。
func (msg MsgBuyName) ValidateBasic() sdk.Error {
	if msg.Buyer.Empty() {
		return sdk.ErrInvalidAddress(msg.Buyer.String())
	}
	if len(msg.Name) == 0 {
		return sdk.ErrUnknownRequest("Name cannot be empty")
	}
	if !msg.Bid.IsAllPositive() {
		return sdk.ErrInsufficientCoins("Bids must be positive")
	}
	return nil
}

// GetSignBytesは署名用のメッセージをエンコードします
func (msg MsgBuyName) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// GetSignersは誰の署名が必要かを定義します
func (msg MsgBuyName) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Buyer}
}
```
次に、`./x/nameservice/handler.go`ファイルで、モジュールルーターに`MsgBuyName`ハンドラーを追加します。

```go
// NewHandlerは "nameservice"タイプのメッセージのハンドラを返します。
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetName:
			return handleMsgSetName(ctx, keeper, msg)
		case MsgBuyName:
			return handleMsgBuyName(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
```

最後に、メッセージによって引き起こされる状態遷移を実行する`BuyName` `handler`関数を定義します。この時点でメッセージは`ValidateBasic`関数を実行しているので、何らかの入力検証があったことを覚えておいてください。しかし、`ValidateBasic`はアプリケーションの状態を問い合わせることはできません。ネットワークの状態に依存する検証ロジック(例えば口座残高)は`handler`関数で実行されるべきです。

```go
//名前を買うためのメッセージを処理する
func handleMsgBuyName(ctx sdk.Context, keeper Keeper, msg MsgBuyName) sdk.Result {
	if keeper.GetPrice(ctx, msg.Name).IsAllGT(msg.Bid) { //入札価格が現在の所有者が支払った価格より高いかどうかを確認します
		return sdk.ErrInsufficientCoins("Bid not high enough").Result() //そうでなければ、エラーをスローします
	}
	if keeper.HasOwner(ctx, msg.Name) {
		_, err := keeper.coinKeeper.SendCoins(ctx, msg.Buyer, keeper.GetOwner(ctx, msg.Name), msg.Bid)
		if err != nil {
			return sdk.ErrInsufficientCoins("Buyer does not have enough coins").Result()
		}
	} else {
		_, _, err := keeper.coinKeeper.SubtractCoins(ctx, msg.Buyer, msg.Bid) //もしそうなら、送り主から入札金額を差し引く
		if err != nil {
			return sdk.ErrInsufficientCoins("Buyer does not have enough coins").Result()
		}
	}
	keeper.SetOwner(ctx, msg.Name, msg.Buyer)
	keeper.SetPrice(ctx, msg.Name, msg.Bid)
	return sdk.Result{}
}
```

まず入札が現在の価格より高いことを確認してください。次に、名前にすでに所有者があるかどうかを確認します。もしそうであれば、前の所有者は「買い手」からお金を受け取る。

所有者がいない場合、あなたの`nameservice`モジュールは`Buyer`からのコインを「燃やし」ます(すなわち回復不可能なアドレスに送ります)。

`SubtractCoins`または`SendCoins`がnil以外のエラーを返すと、ハンドラはエラーをスローして状態遷移を元に戻します。そうでなければ、`Keeper`で以前に定義されたゲッターとセッターを使って、ハンドラーは買い手を新しい所有者に設定し、新しい価格を現在の入札に設定します。

> _*NOTE*_：このハンドラは`coinKeeper`の関数を使って通貨操作を行います。アプリケーションが通貨操作を実行している場合は、[このモジュールのgodocs](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank#BaseKeeper)を参照してください。それが公開する機能。

### あなたの`Msgs`と`Handlers`が定義されたので、これらのトランザクションからデータを作ることについて学ぶ時が来ました[問い合わせに利用可能](08_queriers.md)!

＃メッセージとハンドラ

`Keeper`セットアップができたので、今度は実際にユーザーが名前を購入してそれらに値を設定できるようにする` Msgs`と `Handlers`を構築します。

## `メッセージ`

`Msgs`は状態遷移を引き起こします。 `Msgs`はクライアントがネットワークに送信する[` Txs`](https://github.com/cosmos/cosmos-sdk/blob/develop/types/tx_msg.go#L34-L38)にラップされています。 Cosmos SDKは、 `Txs`から` Msgs`をラップしたりアンラップしたりします。つまり、アプリ開発者としては、Msgsを定義するだけで済みます。 `Msgs`は次のインターフェースを満たさなければなりません(これらはすべて次のセクションで実装します)。

```go
// Transactions messages must fulfill the Msg
type Msg interface {
	// Return the message type.
	// Must be alphanumeric or empty.
	Type() string

	// Returns a human-readable string for the message, intended for utilization
	// within tags
	Route() string

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() Error

	// Get the canonical byte representation of the Msg.
	GetSignBytes() []byte

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []AccAddress
}
```

## `ハンドラ`

`Handlers`は与えられた` Msg`が受信された時にとるべきアクション(どのストアを更新する必要があるか、どのように、そしてどんな条件下で)を定義します。

このモジュールには、ユーザがアプリケーションの状態とやり取りするために送信できる2種類の `Msgs`があります：[` SetName`](set-name.md)と[`BuyName`](./ buy-name.md)です。それらはそれぞれ関連する `Handler`を持ちます。

###これで `Msgs`と` Handlers`の理解が深まったので、最初のメッセージを作り始めることができます：[`SetName`](set-name.md)。

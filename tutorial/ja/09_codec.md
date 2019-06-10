# コーデックファイル

型をAminoに登録する(https://github.com/tendermint/go-amam#registering-types)ようにするために、それらをエンコード/デコードできるようにするには、`/x/nameservice/codec.go`に入れる必要があるコードが少しあります。あなたが作成したインターフェースやインターフェースを実装する構造体は`RegisterCodec`関数で宣言する必要があります。このモジュールでは、2つの`Msg`実装(`SetName`と`BuyName`)を登録する必要がありますが、あなたの`Whois`クエリの戻り型はそうではありません：

```go
package nameservice

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodecはワイヤコーデックに具象型を登録する
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSetName{}, "nameservice/SetName", nil)
	cdc.RegisterConcrete(MsgBuyName{}, "nameservice/BuyName", nil)
}
```

### 次に、あなたは自分のモジュールで[CLI interaction](./10_cli.md)を定義する必要があります。
# アプリケーションを起動します

新しいファイルを作成することから始めましょう： `./app.go`。このファイルはあなたの決定論的ステートマシンの中心です。

`app.go`では、アプリケーションがトランザクションを受け取った時に何をするのかを定義します。しかし最初に、正しい順序でトランザクションを受け取るようにする必要があります。これが[Tendermintコンセンサスエンジン](https://github.com/tendermint/tendermint)の役割です。

必要な依存関係をインポートすることから始めましょう。
```go
package app

import (
  "github.com/tendermint/tendermint/libs/log"
  "github.com/cosmos/cosmos-sdk/x/auth"

  bam "github.com/cosmos/cosmos-sdk/baseapp"
  dbm "github.com/tendermint/tendermint/libs/db"
)
```

インポートされた各モジュールとパッケージのgodocsへのリンクです：

 -  [`log`](https://godoc.org/github.com/tendermint/tendermint/libs/log)：Tendermintのロガー。
 -  [`auth`](https://godoc.org/github.com/cosmos/cosmos-sdk/x/auth)：Comsos SDK用の` auth`モジュール。
 -  [`dbm`](https://godoc.org/github.com/tendermint/tendermint/libs/db)：Tendermintデータベースを操作するためのコード。
 -  [`baseapp`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp)：下記を参照してください

ここにあるパッケージのいくつかは `tendermint`パッケージです。 Tendermintは[ABCI](https://github.com/tendermint/tendermint/tree/master/abci)と呼ばれるインタフェースを通してネットワークからアプリケーションにトランザクションを渡します。構築しているブロックチェーンノードのアーキテクチャを見ると、次のようになっています。

```
+---------------------+
|                     |
|     Application     |
|                     |
+--------+---+--------+
         ^   |
         |   | ABCI
         |   v
+--------+---+--------+
|                     |
|                     |
|     Tendermint      |
|                     |
|                     |
+---------------------+
```

幸い、ABCIインターフェースを実装する必要はありません。 Cosmos SDKはそれのボイラープレートを[`baseapp`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp)の形で提供します。

これが `baseapp`がすることです：

 - Tendermintコンセンサスエンジンから受け取ったトランザクションをデコードします。
 - トランザクションからメッセージを抽出し、基本的な健全性チェックを行います。
 - メッセージを処理できるように適切なモジュールにルーティングします。 `baseapp`はあなたが使いたい特定のモジュールについての知識を持っていないことに注意してください。このチュートリアルの後半で見るように、そのようなモジュールを `app.go`で宣言するのはあなたの仕事です。 `baseapp`はどのモジュールにも適用できるコアルーティングロジックのみを実装しています。
 -  ABCIメッセージが[`DeliverTx`](https://tendermint.com/docs/spec/abci/abci.html#delivertx)([` CheckTx`](https://tendermint.com/docs/)の場合にコミットspec/abci/abci.html# checktx)変更は永続的ではありません)。
 -  [`Beginblock`](https://tendermint.com/docs/spec/abci/abci.html#beginblock)および[` Endblock`](https://tendermint.com/docs/spec/abci)のセットアップを手伝ってください/abci.html#endblock)は、各ブロックの最初と最後に実行されるロジックを定義するための2つのメッセージです。実際には、各モジュールはそれぞれ独自の `BeginBlock`と` EndBlock`サブロジックを実装しています。アプリの役割はすべてをまとめることです(_注意：あなたのアプリケーションではこれらのメッセージを使わないでしょう)。
 - あなたの状態を初期化するのに役立ちます。
 - クエリ設定に役立ちます。

今度はあなたのアプリケーション用に新しいカスタム型 `nameServiceApp`を作成する必要があります。この型は `baseapp`を埋め込む(他の言語の継承と同様にGoに埋め込む)、つまり`baseapp`のすべてのメソッドにアクセスできるようになります。

```go
const (
    appName = "nameservice"
)

type nameServiceApp struct {
    *bam.BaseApp
}
```

アプリケーションにシンプルなコンストラクタを追加します。

```go
func NewNameServiceApp(logger log.Logger, db dbm.DB) *nameServiceApp {

    // 最初に、さまざまなモジュールで共有されるトップレベルのコーデックを定義します。注：コーデックについては後で説明します
    cdc := MakeCodec()

    // BaseAppは、ABCIプロトコルを通じてTendermintとのやり取りを処理します。
    bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

    var app = &nameServiceApp{
        BaseApp: bApp,
        cdc:     cdc,
    }

    return app
}
```

すばらしい！これであなたはあなたのアプリケーションの骨格を持っています。しかしながら、それはまだ機能性を欠いています。

`baseapp`はあなたがあなたのアプリケーションで使いたいルートやユーザーインタラクションについての知識を持っていません。アプリケーションの主な役割はこれらの経路を定義することです。もう1つの役割は、初期状態を定義することです。これらのことは両方ともあなたがあなたのアプリケーションにモジュールを追加することを必要とします。

[application design](./ app-design.md)セクションで見たように、ネームサービスには3つのモジュールが必要です。 `auth`、` bank`、そして `nameservice`です。最初の2つはすでに存在しますが、最後は存在しません。 `nameservice`モジュールはあなたのステートマシンの大部分を定義します。次のステップはそれを構築することです。

###あなたのアプリケーションを完成させるためには、モジュールを含める必要があります。先に進んで[あなたのネームサービスモジュールの構築を始めましょう](types.md)。あなたは後で `app.go`に戻るでしょう。
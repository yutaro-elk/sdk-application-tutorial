# はじめに

このチュートリアルでは、機能的な[Cosmos SDK](https://github.com/cosmos/cosmos-sdk/)アプリケーションを作成し、その過程でSDKの基本概念と構造を学びます。この例では、Cosmos SDKを利用して**迅速かつ簡単に独自のブロックチェーンを構築する**方法を紹介します。

このチュートリアルの終わりまでに、あなたは機能的な `nameservice`アプリケーション、他の文字列への文字列のマッピング(` map [string] string`)を持つことになります。これは[Namecoin](https://namecoin.org/)、[ENS](https://ens.domains/)、または[Handshake](https://handshake.org/)の 。伝統的なDNSシステム( `map [domain] zonefile`)。ユーザーは未使用の名前を購入したり、自分の名前を売買することができます。

このチュートリアルプロジェクトの最終的なソースコードはすべてこのディレクトリにあります(そしてコンパイルされます)。ただし、手動で作業を進め、自分でプロジェクトを構築してみることをお勧めします。

##必要条件

 -  [`golang`> 1.12.1](https://golang.org/doc/install)インストール済み
 - 動いている[`$ GOPATH`](https://github.com/golang/go/wiki/SettingGOPATH)
 - あなた自身のブロックチェーンを作りたい！

##チュートリアル

このチュートリアルでは、アプリケーションを構成する次のファイルを作成します。

```bash
./nameservice
├──Gopkg.toml
├──Makefile
├──app.go
├──cmd
│├──nscli
││└──main.go
│└──nsd
│└──main.go
└──x
    └──ネームサービス
        ├──クライアント
        │├──cli
        ││├──query.go
        ││└──tx.go
        │├──rest
        ││└──rest.go
        │└──module_client.go
        ├──codec.go
        ├──handler.go
        ├──keeper.go
        ├──msgs.go
        ├──querier.go
        └──types.go
```
新しいgitリポジトリを作成することから始めます。
```bash
mkdir -p $ GOPATH/src/github.com/{。ユーザー名}/nameservice
cd $ GOPATH/src/github.com/{。ユーザー名}/nameservice
git init
```

それでは、ただフォローしてください。最初のステップでは、アプリケーションの設計について説明します。コーディングセクションに直接ジャンプしたい場合は、[2番目のステップ](./keeper.md)から始めることができます。

###チュートリアルパート

1. アプリケーションを[デザイン](./app-design.md)します。
2. [`./app.go`](..app_init.md)でアプリケーションの実装を始めます。
3. いくつかの基本的な[`Types`](types.md)を定義してモジュールの構築を始めます。
4. [`Keeper`](./keeper.md)を使ってモジュールのメインコアを作成します。
5. [`Msgs`と` Handlers`](./msgs-handlers.md)を通して状態遷移を定義します。
* [`SetName`](set-name.md)
* [`BuyName`](./buy-name.md)
6. [`Queriers`](./queriers.md)を使ってあなたのステートマシンのビューを作ります。
7. [`sdk.Codec`](./codec.md)を使ってエンコーディングフォーマットで型を登録します。
8. [あなたのモジュール用のCLIインタラクション](./cli.md)を作成します。
9. [自分のネームサービスにアクセスするためのクライアント用のHTTPルート](rest.md)を作成します。
10. モジュールをインポートして[アプリケーションのビルドを終了します](./app-complete.md)！
11. アプリケーションに[`nsd`と` nscli`エントリポイント](./entrypoint.md)を作成します。
12. [`dep`を使った依存関係管理](./dep.md)を設定します。
13. 例を[ビルドして実行](./build-run.md)します。
14. [RESTルートを実行する](run-rest.md)。

##チュートリアルを始めるために[ここをクリック](./app-design.md)
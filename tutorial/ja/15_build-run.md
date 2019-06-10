# アプリケーションの構築と実行

## `nameservice`アプリケーションを構築する

機能を確認するためにこのリポジトリで`nameservice`アプリケーションを構築したい場合は**Go 1.12.1 +**が必要です。

これまでに`go mod`を使ったことがないのであれば、environmentにいくつかのパラメータを追加する必要があります。

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bash_profile
echo "export GOBIN=\$GOPATH/bin" >> ~/.bash_profile
echo "export PATH=\$PATH:\$GOBIN" >> ~/.bash_profile
echo "export GO111MODULE=on" >> ~/.bash_profile
source ~/.bash_profile
```

これで、アプリケーションをインストールして実行することができます。

```bash
# あなたの$GOBINにアプリをインストールする
make install

# これで、以下のコマンドを実行できるはずです。
nsd help
nscli help
```

## ライブネットワークを実行してコマンドを使用する

あなたのアプリケーションのための設定と`genesis.json`ファイルとトランザクションのためのアカウントを初期化するために、実行することから始めてください：

> _*NOTE*_：以下のコマンドでは、アドレスは端末ユーティリティを使って引き出されます。以下に示すように、キーの作成から保存された生の文字列を入力することもできます。コマンドはあなたのマシンに[`jq`](https://stedolan.github.io/jq/download/)がインストールされていることを必要とします。

> _*NOTE*_：以前にチュートリアルを実行したことがあるなら、最初から`nsd unsafe-reset-all`を使うか、両方のホームフォルダ`rm -rf〜/ .ns *`を削除することで始めることができます。

> _*NOTE*_：レジャー用のCosmosアプリを持っていてそれを使いたい場合、`nscli keys add jack`でキーを作成するときに最後に`--ledger`を追加するだけです。必要なものはこれだけです。サインインすると、`jack`は元帳キーとして認識され、デバイスが必要になります。

```bash
# 設定ファイルとgenesisファイルを初期化する
nsd init --chain-id namechain

# ここに`Address`の出力をコピーして後で使うためにそれを保存する
# [オプション]元帳ナノSを使用するには、末尾に "--ledger"を追加します。
nscli keys add jack

# ここに`Address`の出力をコピーして後で使うためにそれを保存する
nscli keys add alice

# コインを使って両方のアカウントをgenesisファイルに追加します。
nsd add-genesis-account $(nscli keys show jack -a) 1000nametoken,1000jackcoin
nsd add-genesis-account $(nscli keys show alice -a) 1000nametoken,1000alicecoin

# chain-idフラグが不要になるようにCLIを設定します。
nscli config chain-id namechain
nscli config output json
nscli config indent true
nscli config trust-node true
```

これで`nsd start`を呼び出して`nsd`を起動することができます。生成中のブロックを表すログのストリーミングが開始されます。これには数秒かかります。

作成したネットワークに対してコマンドを実行するには、別のターミナルを開きます。

```bash
# まず口座をチェックして資金があることを確認します
nscli query account $(nscli keys show jack -a) 
nscli query account $(nscli keys show alice -a) 

# ジェネシスファイルからあなたのコインを使ってあなたのファーストネームを買う
nscli tx nameservice buy-name jack.id 5nametoken --from jack 

# 購入したばかりの名前の値を設定します
nscli tx nameservice set-name jack.id 8.8.8.8 --from jack 

# 登録した名前に対して解決クエリを試してください。
nscli query nameservice resolve jack.id
# > 8.8.8.8

# 登録したばかりの名前に対してwhoisクエリを試してください。
nscli query nameservice whois jack.id
# > {"値"： "8.8.8.8"、 "所有者"： "cosmos1l7k5tdt2qam0zecxrx78yuw447ga54dsmtpk2s"、 "価格"：[{"デノーム"： "名前"： "5"}]}

# アリスはジャックから名前を買う
nscli tx nameservice buy-name jack.id 10nametoken --from alice 
```

### おめでとうございます、Cosmos SDKアプリケーションを作成しました。このチュートリアルはこれで完了です。 RESTサーバーを使って同じコマンドを実行する方法を知りたい場合は[ここをクリック](16_run-rest.md)。

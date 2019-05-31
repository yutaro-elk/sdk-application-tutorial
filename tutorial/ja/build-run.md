＃アプリケーションの構築と実行

## `nameservice`アプリケーションを構築する

機能を確認するためにこのリポジトリで `nameservice`アプリケーションを構築したい場合は** Go 1.12.1 + **が必要です。

これまでに `go mod`を使ったことがないのであれば、environmentにいくつかのパラメータを追加する必要があります。

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
# Install the app into your $GOBIN
make install

# Now you should be able to run the following commands:
nsd help
nscli help
```

##ライブネットワークを実行してコマンドを使用する

あなたのアプリケーションのための設定と `genesis.json`ファイルとトランザクションのためのアカウントを初期化するために、実行することから始めてください：

> _ * NOTE * _：以下のコマンドでは、アドレスは端末ユーティリティを使って引き出されます。以下に示すように、キーの作成から保存された生の文字列を入力することもできます。コマンドはあなたのマシンに[`jq`]（https://stedolan.github.io/jq/download/）がインストールされていることを必要とします。

> _ * NOTE * _：以前にチュートリアルを実行したことがあるなら、最初から `nsd unsafe-reset-all`を使うか、両方のホームフォルダ` rm -rf〜/ .ns * `を削除することで始めることができます。

> _ * NOTE * _：レジャー用のCosmosアプリを持っていてそれを使いたい場合、 `nscli keys add jack`でキーを作成するときに最後に` --ledger`を追加するだけです。必要なものはこれだけです。サインインすると、 `jack`は元帳キーとして認識され、デバイスが必要になります。

```bash
# Initialize configuration files and genesis file
nsd init --chain-id namechain

# Copy the `Address` output here and save it for later use 
# [optional] add "--ledger" at the end to use a Ledger Nano S 
nscli keys add jack

# Copy the `Address` output here and save it for later use
nscli keys add alice

# Add both accounts, with coins to the genesis file
nsd add-genesis-account $(nscli keys show jack -a) 1000nametoken,1000jackcoin
nsd add-genesis-account $(nscli keys show alice -a) 1000nametoken,1000alicecoin

# Configure your CLI to eliminate need for chain-id flag
nscli config chain-id namechain
nscli config output json
nscli config indent true
nscli config trust-node true
```

これで `nsd start`を呼び出して` nsd`を起動することができます。生成中のブロックを表すログのストリーミングが開始されます。これには数秒かかります。

作成したネットワークに対してコマンドを実行するには、別の端末を開きます。

```bash
# First check the accounts to ensure they have funds
nscli query account $(nscli keys show jack -a) 
nscli query account $(nscli keys show alice -a) 

# Buy your first name using your coins from the genesis file
nscli tx nameservice buy-name jack.id 5nametoken --from jack 

# Set the value for the name you just bought
nscli tx nameservice set-name jack.id 8.8.8.8 --from jack 

# Try out a resolve query against the name you registered
nscli query nameservice resolve jack.id
# > 8.8.8.8

# Try out a whois query against the name you just registered
nscli query nameservice whois jack.id
# > {"value":"8.8.8.8","owner":"cosmos1l7k5tdt2qam0zecxrx78yuw447ga54dsmtpk2s","price":[{"denom":"nametoken","amount":"5"}]}

# Alice buys name from jack
nscli tx nameservice buy-name jack.id 10nametoken --from alice 
```

###おめでとうございます、Cosmos SDKアプリケーションを作成しました。このチュートリアルはこれで完了です。 RESTサーバーを使って同じコマンドを実行する方法を知りたい場合は[ここをクリック](run-rest.md)。

# hello_spanner

[Google Cloud Spanner](https://cloud.google.com/spanner/) のハンズオンです。

## 環境構築

https://spanner.gcpug.jp/ を見て、GCPUG Shared Spannerを使えるように権限を貰います。
手元で実行する場合はserviceAccountsは空でかまいません。
gcpugSlackIdが必須になってますが、とりあえず、僕が適当に知ってるので、GopherDojo SlackのIDでOKです。

## ハンズオン

### Part 1 SpannerにQueryを実行する

とりあえず、Cloud Consoleから既存のTableに適当にQueryを実行します。

### Part 2 SpannerにDatabase, Tableを作成する

自分のDatabaseを作成します。
SpannerはDatabase作成時にDDLも実行した方が早いです。

### Part 3 GoからSpannerにInsertを実行する

GoからSpannerを触ってみましょう。
Localから実行する場合は `gcloud auth application-default login` を実行して、Application Default Credentialとして、自分のGoogleアカウントが利用されるようにしておきます。

[`gcloud auth application-default login`](https://cloud.google.com/sdk/gcloud/reference/auth/application-default/login) 

### Part 4 GoからSpannerにQueryを実行する

Spannerからデータを取得する方法はいくつかあるので、Goから試してみます。

### Part 5 Interleave Tableを試す

Spanner固有の機能であるInterleave Tableを試してみます。

### Part 6 Sharding Counterを作成する

分散DBは必ずしも同じRowへの更新が早くないため、Counterのようなものを実装する場合、Shardingをしましょう。

## Close

GCPUG Shard Spannerから自分のDBを削除しておしまいです。

```
gcloud spanner databases delete {your slack id} --project gcpug-public-spanner
```
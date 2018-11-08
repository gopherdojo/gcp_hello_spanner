# GoからSpannerにInsertを実行する

## Spanner Clientを作成する

SpannerをGoから利用する場合 `cloud.google.com/go/spanner` packageを利用します。
SpannerはSessionの作成に2secぐらいかかるので、最初にSpannerClientを作成して、基本的にはそのClientを使いまわしていきます。

```
func createClient(ctx context.Context, db string) (*spanner.Client) {
	dataClient, err := spanner.NewClient(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	return dataClient
}
```

## INSERTを行う

SpannerはDMLに対応しているので、 `INSERT文` も実行できますが、今回は元々存在するMutationをApplyするという方法を試してみます。
`spanner.InsertStruct` を利用して、Insert用のMutationを作成して、最後にApplyします。
ApplyはAtomicに実行されるので、複数行のInsertを行いたい場合は、Mutationを複数Applyに渡します。

```
func insert(ctx context.Context, tweet *Tweet, client *spanner.Client) (error) {
	m, err := spanner.InsertStruct("Tweet", tweet)
	if err != nil {
		return err
	}
	ml := []*spanner.Mutation{
		m,
	}
	_, err = client.Apply(ctx, ml)
	return err
}
```

## INSERTされたRowを確認する

```
gcloud spanner databases execute-sql sinmetal-test1 --project gcpug-public-spanner --sql \
"select * from Tweet"

Id                                    Author    CommitedAt                  Content        Count  CreatedAt                    Favos  Sort  UpdatedAt
dfe856d7-9649-4bd8-962e-08c37e515b32  sinmetal  2018-11-07T09:11:28.78843Z  Hello Spanner  0      2018-11-07T09:11:20.033818Z  []     0     2018-11-07T09:11:20.033818Z
```
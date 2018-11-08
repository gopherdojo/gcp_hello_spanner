# Part 4 GoからSpannerにQueryを実行する

## Keyを指定してのGetを行う

SpannerはKVSにRDBのエッセンスを加えたDBなので、Keyを指定してGETをするという基本的なAPIがあります。
実際にアプリケーションを構築する場合はSQLを主として考えても、もちろん問題ありませんが、基本的な操作として抑えておきますましょう。


以下のサンプルは受け取ったKeyを元にカラムを指定して取得しています。
INSERTを行ったIDを指定して、実際に取得できることを確認してみましょう。

```
func Read(ctx context.Context, ids []string, client *spanner.Client) ([]*Tweet, error) {
	ksl := []spanner.KeySet{}
	for _, v := range ids {
		ksl = append(ksl, spanner.Key{v})
	}

	iter := client.Single().Read(ctx, "Tweet", spanner.KeySets(ksl...), []string{"Id", "Author", "Content", "Count"})
	defer iter.Stop()

	tl := []*Tweet{}
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var t Tweet
		if err := row.ToStruct(&t); err != nil {
			return nil, err
		}
		tl = append(tl, &t)
	}
	return tl, nil
}
```

## 全件取得する

件数が少ないMaster Tableの取得、UnitTestやMigrationでのTableのRowを全件取得したい場合に便利なのが、 `spanner.AllKeys` です。
Keyの一覧の代わりに指定すれば、Tableを全件取得できます。 

```
func ReadAllKeys(ctx context.Context, client *spanner.Client) ([]*Tweet, error) {
	iter := client.Single().Read(ctx, "Tweet", spanner.AllKeys(), []string{"Id", "Author", "Content", "Count"})
	defer iter.Stop()

	tl := []*Tweet{}
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var t Tweet
		if err := row.ToStruct(&t); err != nil {
			return nil, err
		}
		tl = append(tl, &t)
	}
	return tl, nil
}
```

## SimpleにSQLを実行する

ここまではKVSのような指定の方法でしたが、SQLを実行することもできます。
SQLを実行する場合 `spanner.Statement` を利用します。

```
func Query(ctx context.Context, sql string, params map[string]interface{}, client *spanner.Client) ([]*Tweet, error) {
	iter := client.Single().Query(ctx, spanner.Statement{
		SQL:    sql,
		Params: params,
	})
	defer iter.Stop()

	tl := []*Tweet{}
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var t Tweet
		if err := row.ToStruct(&t); err != nil {
			return nil, err
		}
		tl = append(tl, &t)
	}
	return tl, nil
}
```

SQLを実行する時に重要なのが、変数と指定したいパラメータは文字列連結でSQLを組み立てたりするのではなく、`spanner.Statement.Parass` を指定することです。
SpannerはQuery PlanをCacheしますが、その数に限りがあります。
Paramsを使わずに変数値をSQLとして埋め込むと、別々のSQLとしてみなされ、Cacheを有効活用できなくなってしまいます。

```
Query(ctx, "SELECT * FROM Tweet WHERE Count >= @count", map[string]interface{}{
    "count": 0,
}, spannerClient)
```

## Indexを参照してGetを行う

IndexもTableであるを象徴するような機能が `ReadUsingIndex` です。
この機能は指定したIndexを参照してRowを取得することができます。
この時、注意すべき点は `ReadUsingIndex` で取得できるColumnはIndexに含まれているもののみです。
単純にIndexを作っただけの場合は、Indexの値とPKが含まれています。
フィルターやソートには利用しないが、取得したいColumnがある場合は [STORING](https://cloud.google.com/spanner/docs/secondary-indexes#storing_clause) の機能を利用して、IndexにColumnを含めておきます。

```
func ListOrderByCountDesc(ctx context.Context, client *spanner.Client) ([]*Tweet, error) {
	// Countの値が入ったTweetCountDesc Indexに対するKeyなので、Countの値を入れる
	kr := spanner.KeyRange{
		Start: spanner.Key{5},
		End:   spanner.Key{0},
	}

	iter := client.Single().ReadUsingIndex(ctx, "Tweet", "TweetCountDesc", kr, []string{"Id", "Count"})
	defer iter.Stop()

	tl := []*Tweet{}
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var t Tweet
		if err := row.ToStruct(&t); err != nil {
			return nil, err
		}
		tl = append(tl, &t)
	}
	return tl, nil
}
```

## SQLでIndexを利用する

SQLでIndexを利用する場合、 `FORCE_INDEX` を指定します。
Spannerが判断してよしなにIndexを利用してくれる時もあるのですが、大抵使ってくれないので、明示的に指定するのが無難です。

```
SELECT * FROM Tweet@{FORCE_INDEX=TweetCountDesc} WHERE Count >= @count
```

SQLでIndexが利用されたかどうかは、Explanationで確認できます。

![Explanation](https://github.com/sinmetal/hello_spanner/blob/master/part4/spanner_explanation.png)
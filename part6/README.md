# Part 6 Sharding Counterを作成する

Transactionの衝突を避けてデータを更新していくために、複数のRowに分散してデータを保存するSharding Counterを作成してみましょう。

## Tableの作成

カウントを保存するためのAccessCounter Tableを作成します。

```sql
gcloud spanner databases ddl update {your slack id} --project gcpug-public-spanner --ddl \
"CREATE TABLE AccessCounter ( \
  ID STRING(MAX) NOT NULL, \
  Count INT64 NOT NULL, \
  CommitedAt TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true), \
) PRIMARY KEY (ID);"
```

## カウンタをインクリメントする

複数のRowをカウンタ用に作成し、ランダムに1つ選んでインクリメントを行います。
Shardの数を増やせば同時にカウントできる数が増えますが、取得する時の合算が辛くなるので、あまり大きな値にしない方が無難です。

```go
const (
	accessCounterTableName = "AccessCounter"
	accessCounterIDPrefix  = "AccessCounter-"
	accessCounterShardMax  = 10
)

// IncrementAccessCounter is カウンターをインクリメントする
// 返し値のint64はShardingCounter全体の値ではなく、更新したカウンタの値
func IncrementAccessCounter(ctx context.Context, client *spanner.Client) (int64, error) {
	var count int64
	shard := rand.Intn(accessCounterShardMax)
	id := fmt.Sprintf("%s%d", accessCounterIDPrefix, shard)
	key := spanner.Key{id}
	_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		row, err := txn.ReadRow(ctx, accessCounterTableName, key, []string{"Count"})
		if err != nil {
			if spanner.ErrCode(err) != codes.NotFound {
				return err
			}
			// noop
		} else {
			if err := row.Column(0, &count); err != nil {
				return err
			}
		}

		count++
		a := AccessCounter{
			ID:         id,
			Count:      count,
			CommitedAt: spanner.CommitTimestamp,
		}
		m, err := spanner.InsertOrUpdateStruct(accessCounterTableName, a)
		if err != nil {
			return err
		}
		if err := txn.BufferWrite([]*spanner.Mutation{m}); err != nil {
			return err
		}
		return nil
	})
	return count, err
}
```

## カウンタを取得する

AccessCounterのRowを取得して、Goのコード上で合算します。

```go
func GetAccessCounter(ctx context.Context, client *spanner.Client) (int64, error) {
	k := spanner.KeyRange{
		Start: spanner.Key{accessCounterIDPrefix},
		End:   spanner.Key{fmt.Sprintf("%s%d", accessCounterIDPrefix, math.MaxInt64)},
	}
	iter := client.ReadOnlyTransaction().Read(ctx, accessCounterTableName, k, []string{"Count"})
	defer iter.Stop()
	var ac int64
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return ac, err
		}

		var count int64
		if err := row.Column(0, &count); err != nil {
			return ac, err
		}
		ac += count
	}
	return ac, nil
}
```

## 暇だったら

### カウンタの合算をSQLで行う

SpannerのSQLは集計関数も利用できるので、SQLで合算するようにしてみましょう。

### やたらとスケールするSharding Counterを作成する

上記のSharding CounterはKeyのPrefixを元にしているため、同じSplitに保存されている可能性が高いです。
そのため、Splitが過負荷になっている場合Performanceが落ちる可能性があります。
Splitが分散するような仕組みのSharding Counterを作ってみましょう。
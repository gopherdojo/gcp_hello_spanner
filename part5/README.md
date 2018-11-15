# Part 5 Interleave Tableを試す

Spanner固有の概念として [Interleave](https://cloud.google.com/spanner/docs/schema-and-data-model#creating-interleaved-tables) があります。
この機能は関係の深い親子テーブルを同じスプリットに保存する機能です。
ツリー状に階層は作成できますが、複数のツリーに所属することはできません。
PKの値に制約があるため、簡単に変更することもできないため、よく考えて使う機能です。
受注伝票と受注明細のような必ずセットになっているようなデータに対して利用します。

## Interleave Tableを作る

親となる `Orders` Tableと子どもの `OrdersDetail` Tableを作成します。
Interleaveにする場合、子どもは親のPKを自分のPKの先頭に持つ必要があるので、 `OrdersDetail.Id` は `Order.Id` が入ります。
今回のサンプルでは、 `OrdersDetail` のPKは `Id, ItemId` としているので、同じOrderの中に同じItemは一つとなるような設計です。
`ON DELETE CASCADE` も付けているので、 `Orders` Tableが削除されると同時に子どもの `OrdersDetail` も削除されます。

```
gcloud spanner databases ddl update {your slack id} --project gcpug-public-spanner --ddl \
"CREATE TABLE Orders ( \
  Id STRING(MAX) NOT NULL, \
  CustomerEmail STRING(MAX) NOT NULL, \
  CreatedAt TIMESTAMP NOT NULL, \
  UpdatedAt TIMESTAMP NOT NULL, \
  CommitedAt TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true), \
) PRIMARY KEY (Id); \
 \
CREATE TABLE OrdersDetail ( \
  Id STRING(MAX) NOT NULL, \
  ItemId STRING(MAX) NOT NULL, \
  Price INT64 NOT NULL, \
  Count INT64 NOT NULL, \
  CreatedAt TIMESTAMP, \
  UpdatedAt TIMESTAMP NOT NULL, \
  CommitedAt TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true), \
) PRIMARY KEY (Id, ItemId), \
  INTERLEAVE IN PARENT Orders ON DELETE CASCADE;"
```

Interleave とは関係ないですが、それっぽくItem Master Tableも作っておきます。
SpannerにはForeign keyの機能は存在しないので、Item TableとOrdersDetail Tableの間には特に制約はありません。

```
gcloud spanner databases ddl update {your slack id} --project gcpug-public-spanner --ddl \
"CREATE TABLE Item ( \
  Id STRING(MAX) NOT NULL, \
  Name STRING(MAX) NOT NULL, \
  Price INT64 NOT NULL, \
) PRIMARY KEY (Id)"
```

## OrderをInsertする

### ItemをInsertする

```go
type Item struct {
	ID    string `spanner:"Id"`
	Name  string
	Price int64
}

func InsertItem(ctx context.Context, client *spanner.Client) (Item, error) {
	item := Item{ID: uuid.New().String(),
		Name:  "hoge",
		Price: 100,
	}

	_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL:    `INSERT INTO Item (Id, Name, Price) VALUES (@item.ID, @item.Name, @item.Price)`,
			Params: map[string]interface{}{"item": &item},
		}
		_, err := txn.Update(ctx, stmt)
		if err != nil {
			return err
		}
		return nil
	})
	return item, err
}
```

### OrderをInsertする

Orders TableとOrdersDetail Tableを同じTransactionの中でInsertします。
ItemID, Price, Countの値は適当な値でDBの制約としては問題ないですが、Item Tableから値を取得して設定するようにしてもかまいません。

```go
type Orders struct {
	ID            string `spanner:"Id"`
	CustomerEmail string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CommitedAt    time.Time
}

type OrdersDetail struct {
	ID         string `spanner:"Id"`
	ItemID     string `spanner:"ItemId"`
	Price      int64
	Count      int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CommitedAt time.Time
}

type OrdersDetailParam struct {
	ItemID string
	Price  int64
	Count  int64
}

func InsertOrder(ctx context.Context, email string, details []OrdersDetailParam, client *spanner.Client) (*Orders, error) {
	id := uuid.New().String()
	now := time.Now()

	o := Orders{
		ID:            id,
		CustomerEmail: email,
		CreatedAt:     now,
		UpdatedAt:     now,
		CommitedAt:    spanner.CommitTimestamp,
	}
	odl := []OrdersDetail{}
	for _, v := range details {
		od := OrdersDetail{
			ID:         id,
			ItemID:     v.ItemID,
			Price:      v.Price,
			Count:      v.Count,
			CreatedAt:  now,
			UpdatedAt:  now,
			CommitedAt: spanner.CommitTimestamp,
		}
		odl = append(odl, od)
	}

	ml := []*spanner.Mutation{}
	om, err := spanner.InsertStruct("Orders", o)
	if err != nil {
		return nil, err
	}
	ml = append(ml, om)

	for _, v := range odl {
		m, err := spanner.InsertStruct("OrdersDetail", v)
		if err != nil {
			return nil, err
		}
		ml = append(ml, m)
	}

	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		if err := txn.BufferWrite(ml); err != nil {
			return err
		}

		return nil
	})

	return &o, err
}
```

### Ordersを取得する

JOINを利用して、Orders TableとOrdersDetail TableのRowを1回のQueryで取得します。
Goで扱いやすいように、Ordersの中に複数Detailを持った状態にするように `ARRAY_AGG` を利用して、Detailをまとめています。

```go
type OrdersWithDetails struct {
	ID            string `spanner:"Id"`
	CustomerEmail string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CommitedAt    time.Time
	Details       []*OrdersDetail
}

func QueryOrders(ctx context.Context, client *spanner.Client) ([]*OrdersWithDetails, error) {
	sql := `SELECT O.Id, CustomerEmail, ARRAY_AGG(STRUCT(D.ItemId As ItemId, D.Price As Price, D.Count As Count)) As Details FROM Orders O JOIN OrdersDetail D ON O.Id = D.Id GROUP BY O.Id, CustomerEmail`
	iter := client.Single().Query(ctx, spanner.NewStatement(sql))
	defer iter.Stop()

	ol := []*OrdersWithDetails{}
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		o := &OrdersWithDetails{}
		if err := row.ToStruct(o); err != nil {
			return nil, err
		}
		ol = append(ol, o)
	}
	return ol, nil
}
```
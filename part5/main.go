package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/k0kubun/pp"
	"google.golang.org/api/iterator"
)

const databaseName = "{your slack id}"

func main() {
	fmt.Println("ignite")

	ctx := context.Background()

	sc := CreateClient(ctx, fmt.Sprintf("projects/gcpug-public-spanner/instances/merpay-sponsored-instance/databases/%s", databaseName))

	item1, err := InsertItem(ctx, sc)
	if err != nil {
		log.Fatal(err)
	}
	item2, err := InsertItem(ctx, sc)
	if err != nil {
		log.Fatal(err)
	}

	details := []OrdersDetailParam{}
	details = append(details, OrdersDetailParam{
		ItemID: item1.ID,
		Price:  item1.Price,
		Count:  2,
	})
	details = append(details, OrdersDetailParam{
		ItemID: item2.ID,
		Price:  item2.Price,
		Count:  3,
	})
	o, err := InsertOrder(ctx, "hoge@example.com", details, sc)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Order.ID is %s\n", o.ID)

	ol, err := QueryOrders(ctx, sc)
	if err != nil {
		log.Fatal(err)
	}
	_, err = pp.Println(ol)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("done")
}

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

func CreateClient(ctx context.Context, db string) *spanner.Client {
	dataClient, err := spanner.NewClient(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	return dataClient
}

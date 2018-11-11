package main

import (
	"context"
	"log"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
)

const databaseName = "sinmetal-test1"

type Item struct {
	ID    string `spanner:"Id"`
	Name  string
	Price int64
}

func InsertItem(ctx context.Context, client *spanner.Client) error {
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
	return err
}

func CreateClient(ctx context.Context, db string) *spanner.Client {
	dataClient, err := spanner.NewClient(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	return dataClient
}

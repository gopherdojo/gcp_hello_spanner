package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/spanner"
	"github.com/k0kubun/pp"
)

var testClient *spanner.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	// CreateSessionに時間がかかると最初に実行されたテストに時間がかかってるように見えるので、Sessionのウォームアップをしておく
	testClient = CreateClient(ctx, fmt.Sprintf("projects/gcpug-public-spanner/instances/merpay-sponsored-instance/databases/%s", databaseName))
	if err := testClient.Single().Query(ctx, spanner.NewStatement("SELECT 1")).Do(func(*spanner.Row) error { return nil }); err != nil {
		log.Fatalf("failed %+v", err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestInsertItem(t *testing.T) {
	ctx := context.Background()

	if _, err := InsertItem(ctx, testClient); err != nil {
		t.Fatal(err)
	}
}

func TestInsertOrder(t *testing.T) {
	ctx := context.Background()

	details := []OrdersDetailParam{}
	details = append(details, OrdersDetailParam{
		ItemID: "hoge",
		Price:  100,
		Count:  2,
	})
	details = append(details, OrdersDetailParam{
		ItemID: "fuga",
		Price:  200,
		Count:  3,
	})
	if _, err := InsertOrder(ctx, "hoge@example.com", details, testClient); err != nil {
		t.Fatal(err)
	}
}

func TestQueryOrders(t *testing.T) {
	ctx := context.Background()

	ol, err := QueryOrders(ctx, testClient)
	if err != nil {
		t.Fatal(err)
	}
	if ol[0].CustomerEmail == "" {
		t.Errorf("CustomerEmail is Empty")
	}
	if len(ol[0].Details) == 0 {
		t.Errorf("Details is Empty")
	}
	pp.Println(ol)
}

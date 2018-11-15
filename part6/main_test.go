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

func TestIncrement(t *testing.T) {
	ctx := context.Background()

	count, err := IncrementAccessCounter(ctx, testClient)
	if err != nil {
		t.Fatal(err)
	}
	if count < 1 {
		t.Errorf("count is small value, val=%d", count)
	}
	pp.Println(count)
}

func TestGetAccessCounter(t *testing.T) {
	ctx := context.Background()

	count, err := GetAccessCounter(ctx, testClient)
	if err != nil {
		t.Fatal(err)
	}
	if count < 1 {
		t.Errorf("count is small value, val=%d", count)
	}
	pp.Println(count)
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
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

func TestQuery(t *testing.T) {
	ctx := context.Background()

	tl, err := Query(ctx, "SELECT * FROM Tweet WHERE Count >= @count", map[string]interface{}{
		"count": 0,
	}, testClient)
	if err != nil {
		t.Fatal(err)
	}
	for _, t := range tl {
		fmt.Printf("%+v\n", t)
	}
}

func TestQueryUsingIndex(t *testing.T) {
	ctx := context.Background()

	tl, err := Query(ctx, "SELECT * FROM Tweet@{FORCE_INDEX=TweetCountDesc} WHERE Count >= @count", map[string]interface{}{
		"count": 1,
	}, testClient)
	if err != nil {
		t.Fatal(err)
	}
	for _, t := range tl {
		fmt.Printf("%+v\n", t)
	}
}

func TestListOrderByCountDesc(t *testing.T) {
	ctx := context.Background()

	tl, err := ListOrderByCountDesc(ctx, testClient)
	if err != nil {
		t.Fatal(err)
	}
	for _, t := range tl {
		fmt.Printf("%+v\n", t)
	}
}

func TestReadAllKeys(t *testing.T) {
	ctx := context.Background()

	tl, err := ReadAllKeys(ctx, testClient)
	if err != nil {
		t.Fatal(err)
	}
	for _, t := range tl {
		fmt.Printf("%+v\n", t)
	}
}

func TestRead(t *testing.T) {
	ctx := context.Background()

	tweet := Tweet{
		ID:     uuid.New().String(),
		Author: "TestUser",
		Favos:  []string{},
	}

	if err := Insert(ctx, &tweet, testClient); err != nil {
		t.Fatal(err)
	}

	tl, err := Read(ctx, []string{tweet.ID}, testClient)
	if err != nil {
		t.Fatal(err)
	}
	for _, t := range tl {
		fmt.Printf("%+v\n", t)
	}
}

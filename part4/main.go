package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

const databaseName = "sinmetal-test1"

type Tweet struct {
	ID         string `spanner:"Id"`
	Author     string
	Content    string
	Count      int64
	Favos      []string
	Sort       int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CommitedAt time.Time
}

func main() {
	fmt.Println("ignite")

	ctx := context.Background()

	sc := CreateClient(ctx, fmt.Sprintf("projects/gcpug-public-spanner/instances/merpay-sponsored-instance/databases/%s", databaseName))

	now := time.Now()
	t := Tweet{
		ID:         uuid.New().String(),
		Author:     "sinmetal",
		Content:    "Hello Spanner",
		Favos:      []string{},
		CreatedAt:  now,
		UpdatedAt:  now,
		CommitedAt: spanner.CommitTimestamp,
	}
	if err := Insert(ctx, &t, sc); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("insert id = %s\n", t.ID)

	tl, err := Query(ctx, "SELECT * FROM Tweet WHERE Count >= @count", map[string]interface{}{
		"count": 0,
	}, sc)
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range tl {
		fmt.Printf("%+v\n", t)
	}

	fmt.Println("done")
}

func CreateClient(ctx context.Context, db string) *spanner.Client {
	dataClient, err := spanner.NewClient(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	return dataClient
}

func Insert(ctx context.Context, tweet *Tweet, client *spanner.Client) error {
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

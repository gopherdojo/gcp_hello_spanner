# Part 2 SpannerにDatabase, Tableを作成する

## 操作するSpanner Instanceの指定

```
gcloud config set spanner/instance merpay-sponsored-instance
```

## Database, Tableの作成

### Database, Tableの作成

`gcloud spanner databases create` を利用してDatabaseを作成します。
ついでに `-ddl` も指定して、Tableなども作成しておきます。
`{your slack id}` にはGCPUG SlackかGopherDojo SlackのIDを入れてください。

```
gcloud spanner databases create {your slack id} --project gcpug-public-spanner --ddl \
"CREATE TABLE Tweet \
 ( \
   Id         STRING(MAX) NOT NULL, \
   Author     STRING(MAX) NOT NULL, \
   CommitedAt TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp= true), \
   Content    STRING(MAX) NOT NULL, \
   CreatedAt  TIMESTAMP NOT NULL, \
   UpdatedAt  TIMESTAMP NOT NULL, \
 ) PRIMARY KEY (Id); \
  \
 CREATE INDEX TweetCountDesc \
 ON Tweet ( \
   Count DESC \
 ); \
  \
 CREATE INDEX TweetUpdatedAtDesc \
 ON Tweet ( \
   UpdatedAt DESC \
 );"
```

### Databaseの確認

`gcloud spanner databases list` でInstanceのDatabase一覧を確認します。

```
gcloud spanner databases list --project gcpug-public-spanner
```

### Tableの確認

`information_schema.tables` に対してクエリを実行して、Tableができあがっているかを確認します。
`information_schema.tables`　にはTable名以外にも色んな値が含まれているので、 [Information Schema](https://cloud.google.com/spanner/docs/information-schema) を一度確認しておくとよいでしょう。
 
```
gcloud spanner databases execute-sql {your slack id} --project gcpug-public-spanner --sql \
"select t.table_name from information_schema.tables AS t WHERE t.table_catalog = '' and t.table_schema = ''"
```
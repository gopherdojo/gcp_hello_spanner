# Part 1 SpannerにQueryを実行する

`gcpug-public-spanner.merpay-sponsord-instance.sinmetal_benchmark_a` に適当にデータが入ったTableがいくつかあるので、クエリを実行してみる。
Cloud Consoleから実行すればExplanationが確認できるので、どのようにクエリが実行されているか確認する。

![Spanner Explanation](https://github.com/sinmetal/hello_spanner/blob/master/part1/explanation.png)

## Simple Query

```sql
SELECT * FROM Item1K LIMIT 10
```

## JOIN

```sql
SELECT * FROM Order1K O JOIN OrderDetail1K D ON O.OrderID = D.OrderID LIMIT 10
```

## GROUP BY

```sql
SELECT UserId, Sum(Price) As Price FROM Order1K GROUP BY 1 LIMIT 10
```

## FORCE_INDEX

```sql
SELECT * FROM OrderDetail1K@{FORCE_INDEX=OrderDetail1KItemIdAsc}
WHERE ItemId = "001d465a-8c67-4ba7-9f55-6b2fcf966a13" LIMIT 100
```

![Index Scan](https://github.com/sinmetal/hello_spanner/blob/master/part1/force_index_sample.png)

FORCE_INDEXを外すとIndexを使わなくなるので、TableをFull Scanしてしまう。

```sql
SELECT * FROM OrderDetail1K
WHERE ItemId = "001d465a-8c67-4ba7-9f55-6b2fcf966a13" LIMIT 100
```

![Index Scan](https://github.com/sinmetal/hello_spanner/blob/master/part1/fullscan_sample.png)
CREATE TABLE AccessCounter (
  ID STRING(MAX) NOT NULL,
  Count INT64 NOT NULL,
  CommitedAt TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (ID);


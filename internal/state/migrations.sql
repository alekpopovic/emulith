-- emulith migration 1
CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY);
CREATE TABLE IF NOT EXISTS s3_buckets (name TEXT PRIMARY KEY, region TEXT NOT NULL, created_at TIMESTAMP NOT NULL);
CREATE TABLE IF NOT EXISTS s3_objects (
  bucket TEXT NOT NULL, key TEXT NOT NULL, etag TEXT NOT NULL, size INTEGER NOT NULL,
  content_type TEXT, last_modified TIMESTAMP NOT NULL, body_path TEXT NOT NULL,
  PRIMARY KEY (bucket, key), FOREIGN KEY (bucket) REFERENCES s3_buckets(name) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS sqs_queues (
  name TEXT PRIMARY KEY, url_path TEXT NOT NULL, visibility_timeout_seconds INTEGER NOT NULL, created_at TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS sqs_messages (
  id TEXT PRIMARY KEY, queue_name TEXT NOT NULL, body TEXT NOT NULL, md5 TEXT NOT NULL,
  receipt_handle TEXT, visible_at TIMESTAMP NOT NULL, created_at TIMESTAMP NOT NULL,
  FOREIGN KEY (queue_name) REFERENCES sqs_queues(name) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_sqs_messages_visible ON sqs_messages(queue_name, visible_at, created_at);
INSERT OR IGNORE INTO schema_version(version) VALUES (1);

-- emulith migration 2: DynamoDB durable JSON payloads and canonical binary keys
CREATE TABLE IF NOT EXISTS dynamodb_tables (
  name TEXT PRIMARY KEY, table_id TEXT NOT NULL UNIQUE, arn TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL, created_at TIMESTAMP NOT NULL, billing_mode TEXT NOT NULL,
  partition_key TEXT NOT NULL, partition_type TEXT NOT NULL,
  sort_key TEXT, sort_type TEXT
);
CREATE TABLE IF NOT EXISTS dynamodb_attributes (
  table_name TEXT NOT NULL, name TEXT NOT NULL, attribute_type TEXT NOT NULL,
  PRIMARY KEY(table_name, name), FOREIGN KEY(table_name) REFERENCES dynamodb_tables(name) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS dynamodb_items (
  table_name TEXT NOT NULL, primary_key BLOB NOT NULL, partition_key BLOB NOT NULL,
  sort_key BLOB, payload BLOB NOT NULL, item_size INTEGER NOT NULL, updated_at TIMESTAMP NOT NULL,
  PRIMARY KEY(table_name, primary_key), FOREIGN KEY(table_name) REFERENCES dynamodb_tables(name) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_dynamodb_items_query ON dynamodb_items(table_name, partition_key, sort_key);
INSERT OR IGNORE INTO schema_version(version) VALUES (2);

-- emulith migration 3: SNS topics and future subscription metadata
CREATE TABLE IF NOT EXISTS sns_topics(name TEXT PRIMARY KEY, arn TEXT NOT NULL UNIQUE, display_name TEXT NOT NULL, created_at TIMESTAMP NOT NULL);
CREATE TABLE IF NOT EXISTS sns_subscriptions(id TEXT PRIMARY KEY, topic_arn TEXT NOT NULL, protocol TEXT NOT NULL, endpoint TEXT NOT NULL, raw_delivery INTEGER NOT NULL DEFAULT 0, created_at TIMESTAMP NOT NULL, FOREIGN KEY(topic_arn) REFERENCES sns_topics(arn) ON DELETE CASCADE, UNIQUE(topic_arn,protocol,endpoint));
CREATE INDEX IF NOT EXISTS idx_sns_topics_arn ON sns_topics(arn);
INSERT OR IGNORE INTO schema_version(version) VALUES (3);

CREATE TABLE IF NOT EXISTS cw_log_groups(name TEXT PRIMARY KEY, created_at INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS cw_log_streams(group_name TEXT NOT NULL, name TEXT NOT NULL, created_at INTEGER NOT NULL, PRIMARY KEY(group_name,name), FOREIGN KEY(group_name) REFERENCES cw_log_groups(name) ON DELETE CASCADE);
CREATE TABLE IF NOT EXISTS cw_log_events(id INTEGER PRIMARY KEY AUTOINCREMENT, group_name TEXT NOT NULL, stream_name TEXT NOT NULL, timestamp_ms INTEGER NOT NULL, message TEXT NOT NULL, ingested_ms INTEGER NOT NULL, FOREIGN KEY(group_name,stream_name) REFERENCES cw_log_streams(group_name,name) ON DELETE CASCADE);
CREATE INDEX IF NOT EXISTS idx_cw_log_events ON cw_log_events(group_name,stream_name,timestamp_ms,id);
-- Azure Blob metadata
CREATE TABLE IF NOT EXISTS azure_containers(account TEXT NOT NULL, name TEXT NOT NULL, etag TEXT NOT NULL, last_modified TIMESTAMP NOT NULL, metadata TEXT NOT NULL DEFAULT '{}', created_at TIMESTAMP NOT NULL, PRIMARY KEY(account,name));
CREATE TABLE IF NOT EXISTS azure_blobs(account TEXT NOT NULL, container TEXT NOT NULL, name TEXT NOT NULL, etag TEXT NOT NULL, last_modified TIMESTAMP NOT NULL, body_path TEXT NOT NULL, size INTEGER NOT NULL, content_type TEXT NOT NULL DEFAULT '', content_encoding TEXT NOT NULL DEFAULT '', content_language TEXT NOT NULL DEFAULT '', cache_control TEXT NOT NULL DEFAULT '', content_disposition TEXT NOT NULL DEFAULT '', content_md5 TEXT NOT NULL DEFAULT '', metadata TEXT NOT NULL DEFAULT '{}', created_at TIMESTAMP NOT NULL, PRIMARY KEY(account,container,name), FOREIGN KEY(account,container) REFERENCES azure_containers(account,name) ON DELETE CASCADE);
INSERT OR IGNORE INTO schema_version(version) VALUES (4);

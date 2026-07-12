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

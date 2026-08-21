CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS events (
 id uuid PRIMARY KEY, tenant_id text NOT NULL, aggregate_type text NOT NULL, aggregate_id text NOT NULL,
 aggregate_version bigint NOT NULL, event_type text NOT NULL, payload jsonb NOT NULL, metadata jsonb NOT NULL DEFAULT '{}',
 occurred_at timestamptz NOT NULL DEFAULT now(), global_position bigserial NOT NULL, idempotency_key text NOT NULL,
 UNIQUE(tenant_id,aggregate_type,aggregate_id,aggregate_version), UNIQUE(tenant_id,idempotency_key)
);
CREATE TABLE IF NOT EXISTS idempotency_records (tenant_id text NOT NULL, idempotency_key text NOT NULL, aggregate_type text NOT NULL, aggregate_id text NOT NULL, first_version bigint NOT NULL, event_count int NOT NULL, created_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(tenant_id,idempotency_key));
CREATE INDEX IF NOT EXISTS events_stream_idx ON events(tenant_id,aggregate_type,aggregate_id,aggregate_version);
CREATE INDEX IF NOT EXISTS events_position_idx ON events(tenant_id,global_position);
CREATE TABLE IF NOT EXISTS aggregate_heads (tenant_id text NOT NULL, aggregate_type text NOT NULL, aggregate_id text NOT NULL, version bigint NOT NULL, updated_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(tenant_id,aggregate_type,aggregate_id));
CREATE TABLE IF NOT EXISTS outbox_tasks (id bigserial PRIMARY KEY, tenant_id text NOT NULL, event_id uuid NOT NULL REFERENCES events(id), projection text NOT NULL, status text NOT NULL DEFAULT 'pending', attempts int NOT NULL DEFAULT 0, run_after timestamptz NOT NULL DEFAULT now(), locked_until timestamptz, lease_token uuid, last_error text, UNIQUE(event_id,projection));
ALTER TABLE outbox_tasks ADD COLUMN IF NOT EXISTS lease_token uuid;
CREATE INDEX IF NOT EXISTS outbox_ready_idx ON outbox_tasks(status,run_after);
CREATE TABLE IF NOT EXISTS projection_checkpoints (tenant_id text NOT NULL, projection text NOT NULL, position bigint NOT NULL DEFAULT 0, status text NOT NULL DEFAULT 'idle', last_error text, updated_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(tenant_id,projection));
CREATE TABLE IF NOT EXISTS order_views (tenant_id text NOT NULL, order_id text NOT NULL, buyer_id text NOT NULL, currency text NOT NULL, amount bigint NOT NULL, status text NOT NULL, version bigint NOT NULL, last_event_id uuid NOT NULL, PRIMARY KEY(tenant_id,order_id));
CREATE TABLE IF NOT EXISTS account_views (tenant_id text NOT NULL, account_id text NOT NULL, currency text NOT NULL, balance bigint NOT NULL, credited bigint NOT NULL, debited bigint NOT NULL, version bigint NOT NULL, last_event_id uuid NOT NULL, PRIMARY KEY(tenant_id,account_id));
CREATE TABLE IF NOT EXISTS replay_jobs (id uuid PRIMARY KEY, tenant_id text NOT NULL, projection text NOT NULL, from_position bigint NOT NULL DEFAULT 0, to_position bigint, status text NOT NULL, processed bigint NOT NULL DEFAULT 0, error text, created_at timestamptz NOT NULL DEFAULT now(), finished_at timestamptz);
CREATE TABLE IF NOT EXISTS consistency_checks (id uuid PRIMARY KEY, tenant_id text NOT NULL, aggregate_type text NOT NULL, aggregate_id text NOT NULL, status text NOT NULL, expected jsonb, actual jsonb, detail text, created_at timestamptz NOT NULL DEFAULT now());

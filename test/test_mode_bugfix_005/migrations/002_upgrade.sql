ALTER TABLE outbox_tasks ADD COLUMN IF NOT EXISTS lease_token uuid;
CREATE TABLE IF NOT EXISTS idempotency_records (tenant_id text NOT NULL,idempotency_key text NOT NULL,aggregate_type text NOT NULL,aggregate_id text NOT NULL,first_version bigint NOT NULL,event_count int NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(tenant_id,idempotency_key));

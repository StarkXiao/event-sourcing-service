# Event Sourcing Service

PostgreSQL-backed Go event-sourcing service with optimistic aggregate versions, durable event records, asynchronous projections, replay job tracking, and projection consistency checks.

## Commands

```bash
go build ./...
go test ./...
go run ./cmd/server
docker compose up -d --build
```

The service uses `DATABASE_URL`, `HTTP_ADDR`, `ADMIN_TOKEN`, `AUTO_MIGRATE`, and `WORKER_CONCURRENCY`. See `.env.example` for local defaults.

## Verification

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

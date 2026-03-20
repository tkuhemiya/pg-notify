# Demo Postgres (small image)

This uses a minimal image based on `postgres:17-alpine` and auto-loads schema + demo activity SQL.

## 1) Build image

```bash
docker build -t pg-notify-db:local -f db/Dockerfile db
```

## 2) Run database

```bash
docker run --name pg-notify-db \
  -e POSTGRES_DB=pg_notify \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  -d pg-notify-db:local
```

Set your app config `database_url` to:

```text
postgres://postgres:postgres@localhost:5432/pg_notify?sslmode=disable
```

## 3) Generate demo activity

Single burst:

```bash
docker exec -i pg-notify-db psql -U postgres -d pg_notify -c "SELECT simulate_orders(500);"
```

Continuous load:

```bash
./db/simulate-load.sh pg-notify-db 50 1
```

Arguments are: `container_name batch_size sleep_seconds`.

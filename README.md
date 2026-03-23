# pg-notify

Dashboard + SSE feed for visualizing latency and rate metrics from PostgreSQL
channels you `LISTEN` on.

![pg-notify dashboard screenshot](.github/assets/demo.png)


## Run Against Your Own Database

1. **Make sure your database emits `NOTIFY` events.**

   The app listens on PostgreSQL channels and expects JSON payloads with a
   `created_at` field formatted as RFC3339 or RFC3339Nano.

   Example trigger function:

   ```sql
   CREATE OR REPLACE FUNCTION notify_row_change()
   RETURNS trigger
   LANGUAGE plpgsql
   AS $$
   DECLARE
     ts timestamptz;
   BEGIN
     ts := COALESCE(NEW.created_at, OLD.created_at, NOW());

     PERFORM pg_notify(
       'channel_name',
       json_build_object(
         'created_at',
         to_char(ts, 'YYYY-MM-DD"T"HH24:MI:SS.USOF')
       )::text
     );

     RETURN COALESCE(NEW, OLD);
   END;
   $$;
   ```

   Attach it to a table:

   ```sql
   CREATE TRIGGER my_table_notify_trigger
   AFTER INSERT OR UPDATE OR DELETE ON my_table
   FOR EACH ROW
   EXECUTE FUNCTION notify_row_change();
   ```

2. **Create a `pg-notify.cfg`.**

   ```json
   {
     "database_url": "postgres://user:pass@your-host:5432/your_db?sslmode=require",
     "port": 8080,
     "event_names": ["channel_name", "channel_name2"]
   }
   ```

3. **Run the service.**

   ```bash
   go run .
   ```

   Or build it first:

   ```bash
   go build
   ./pg-notify
   ```

4. **Open the dashboard.**

   Visit:

   ```text
   http://localhost:<port>
   ```

## Demo

For demoing the repository includes a small Postgres
image and helper scripts in `db/`.

Follow `db/README.md` to build and run the container, then run:
```bash
make demo
```
This starts the app and pushes synthetic activity into the demo database.

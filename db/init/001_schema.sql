CREATE TABLE IF NOT EXISTS orders (
  id           BIGSERIAL PRIMARY KEY,
  customer_id  BIGINT NOT NULL,
  status       TEXT   NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION notify_order_insert() RETURNS trigger AS $$
DECLARE
  payload TEXT;
BEGIN
  payload := json_build_object(
    'event', 'order_inserted',
    'order_id', NEW.id,
    'customer_id', NEW.customer_id,
    'created_at', NEW.created_at
  )::text;

  PERFORM pg_notify('orders_inserted', payload);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_notify_order_insert ON orders;
CREATE TRIGGER trg_notify_order_insert
AFTER INSERT ON orders
FOR EACH ROW
EXECUTE FUNCTION notify_order_insert();

CREATE TABLE IF NOT EXISTS shipments (
  id           BIGSERIAL PRIMARY KEY,
  order_id     BIGINT   NOT NULL,
  status       TEXT     NOT NULL,
  carrier      TEXT     NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION notify_shipment_insert() RETURNS trigger AS $$
DECLARE
  payload TEXT;
BEGIN
  payload := json_build_object(
    'event', 'shipment_created',
    'shipment_id', NEW.id,
    'order_id', NEW.order_id,
    'status', NEW.status,
    'carrier', NEW.carrier,
    'created_at', NEW.created_at
  )::text;

  PERFORM pg_notify('shipments_created', payload);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_notify_shipment_insert ON shipments;
CREATE TRIGGER trg_notify_shipment_insert
AFTER INSERT ON shipments
FOR EACH ROW
EXECUTE FUNCTION notify_shipment_insert();

CREATE OR REPLACE FUNCTION simulate_orders(batch_size integer DEFAULT 100)
RETURNS void AS $$
BEGIN
  INSERT INTO orders (customer_id, status, created_at)
  SELECT
    (10000 + floor(random() * 90000))::bigint,
    (ARRAY['pending', 'processing', 'completed', 'failed'])[1 + floor(random() * 4)::int],
    now()
  FROM generate_series(1, GREATEST(batch_size, 1));
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION simulate_shipments(batch_size integer DEFAULT 50)
RETURNS void AS $$
BEGIN
  INSERT INTO shipments (order_id, status, carrier, created_at)
  SELECT
    (10000 + floor(random() * 90000))::bigint,
    (ARRAY['scheduled', 'in_transit', 'delivered', 'delayed'])[1 + floor(random() * 4)::int],
    (ARRAY['UPS', 'FedEx', 'USPS', 'DHL'])[1 + floor(random() * 4)::int],
    now()
  FROM generate_series(1, GREATEST(batch_size, 1));
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION simulate_activity(batch_size integer DEFAULT 100)
RETURNS void AS $$
BEGIN
  PERFORM simulate_orders(GREATEST(batch_size, 1));
  PERFORM simulate_shipments(GREATEST((batch_size + 1) / 2, 1));
END;
$$ LANGUAGE plpgsql;

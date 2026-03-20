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

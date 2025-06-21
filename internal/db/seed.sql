
INSERT INTO warehouses (name, lat, lon)
SELECT
    'Warehouse ' || i,
    28.6139 + (random() - 0.5) * 0.05,
    77.2088 + (random() - 0.5) * 0.05
FROM generate_series(1, 10) AS s(i);

INSERT INTO agents (name, warehouse_id, checked_in_at)
SELECT
    'Agent ' || ((w-1)*20 + a),
    w,
    NOW()::date + INTERVAL '8 hours'
FROM generate_series(1, 10) AS s(w), generate_series(1, 20) AS t(a);

INSERT INTO orders (warehouse_id, lat, lon, delivery_address, scheduled_for, assigned)
SELECT
    w,
    28.6139 + (random() - 0.5) * 0.1,
    77.2088 + (random() - 0.5) * 0.1,
    'Address ' || ((w-1)*60 + o),
    CURRENT_DATE,
    FALSE
FROM generate_series(1, 10) AS s(w), generate_series(1, 60) AS t(o);
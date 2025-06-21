-- Warehouses
CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    lat DOUBLE PRECISION,
    lon DOUBLE PRECISION
);

-- Agents
CREATE TABLE agents (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    warehouse_id INTEGER REFERENCES warehouses(id),
    checked_in_at TIMESTAMP
);

-- Orders
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    warehouse_id INTEGER REFERENCES warehouses(id),
    lat DOUBLE PRECISION,
    lon DOUBLE PRECISION,
    delivery_address TEXT,
    scheduled_for DATE DEFAULT CURRENT_DATE,
    assigned BOOLEAN DEFAULT FALSE
);

-- Assignments
CREATE TABLE agent_assignments (
    id SERIAL PRIMARY KEY,
    agent_id INTEGER REFERENCES agents(id),
    order_id INTEGER REFERENCES orders(id),
    assigned_on DATE DEFAULT CURRENT_DATE,
    distance_km DOUBLE PRECISION,
    estimated_time_minutes INTEGER
);

-- Agent Payouts
CREATE TABLE IF NOT EXISTS agent_payouts (
    id SERIAL PRIMARY KEY,
    agent_id INTEGER NOT NULL REFERENCES agents(id),
    date DATE NOT NULL,
    total_orders INTEGER NOT NULL,
    total_distance DOUBLE PRECISION NOT NULL,
    total_pay DOUBLE PRECISION NOT NULL,
    CONSTRAINT unique_agent_date UNIQUE (agent_id, date)
);


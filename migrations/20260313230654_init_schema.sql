-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT NOT NULL UNIQUE,
  hashed_password TEXT NOT NULL,
  coins NUMERIC(10, 2) DEFAULT 1000 NOT NULL CHECK (coins >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  price NUMERIC(10, 2) NOT NULL CHECK (price >= 0)
);

CREATE TABLE IF NOT EXISTS user_items (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  PRIMARY KEY (user_id, item_id)
);

INSERT INTO items (name, price) VALUES
('t-shirt', 80.00),
('cup', 20.00),
('book', 50.00),
('pen', 10.00),
('powerbank', 200.00),
('hoodie', 300.00),
('umbrella', 200.00),
('socks', 10.00),
('wallet', 50.00),
('pink-hoodie', 500.00);


CREATE TABLE IF NOT EXISTS transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  coins NUMERIC(10, 2) NOT NULL CHECK (coins >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()

  CHECK (from_user_id != to_user_id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_from_user_id
  ON transactions (from_user_id);

CREATE INDEX IF NOT EXISTS idx_transactions_to_user_id
  ON transactions (to_user_id);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS user_items;
DROP TABLE IF EXISTS transactions;
CREATE TABLE products (
  id INT8 PRIMARY KEY DEFAULT unique_rowid(),
  name STRING NOT NULL,
  description STRING,
  price FLOAT8 NOT NULL,
  currency STRING NOT NULL CHECK (char_length(currency) = 3),
  stock_quantity INT4 NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);



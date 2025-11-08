-- name: ListProducts :many
SELECT * FROM products
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (id, name, description, price, currency, stock_quantity)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

-- name: FindProductWithStockInfo :one
SELECT 
  p.id                AS product_id,
  p.name              AS product_name,
  p.description       AS product_description,
  p.price             AS product_price,
  p.currency          AS product_currency,
  p.stock_quantity    AS product_stock_quantity,
  p.created_at        AS product_created_at,
  p.updated_at        AS product_updated_at
FROM products p
WHERE p.id = $1;

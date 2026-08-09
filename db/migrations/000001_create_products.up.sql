CREATE TABLE products (
  id    BIGSERIAL PRIMARY KEY,
  name  VARCHAR(20) NOT NULL,
  price BIGINT NOT NULL CHECK (price >= 0)
);

INSERT INTO products (name, price) VALUES
  ('商品A', 1000),
  ('商品B', 2000),
  ('商品C', 3000);

COMMENT ON TABLE products IS '商品マスタ';
COMMENT ON COLUMN products.id IS '商品ID';
COMMENT ON COLUMN products.name IS '商品名（最大20文字）';
COMMENT ON COLUMN products.price IS '価格（円・0以上）';

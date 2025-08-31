-- 000002_create_order_tables.up.sql

CREATE TABLE IF NOT EXISTS orders (
    order_uid          VARCHAR(255) PRIMARY KEY,
    track_number       VARCHAR(255) NOT NULL,
    entry              VARCHAR(255) NOT NULL,
    locale             VARCHAR(8)   NOT NULL,
    internal_signature TEXT,
    customer_id        VARCHAR(255) NOT NULL,
    delivery_service   VARCHAR(255) NOT NULL,
    shardkey           VARCHAR(255) NOT NULL,
    sm_id              INT          NOT NULL,
    date_created       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    oof_shard          VARCHAR(32)  NOT NULL
);

CREATE TABLE IF NOT EXISTS deliveries (
    id        SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) UNIQUE NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    name      VARCHAR(255) NOT NULL,
    phone     VARCHAR(32)  NOT NULL,
    zip       VARCHAR(32)  NOT NULL,
    city      VARCHAR(255) NOT NULL,
    address   VARCHAR(255) NOT NULL,
    region    VARCHAR(255) NOT NULL,
    email     VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
    id            SERIAL PRIMARY KEY,
    order_uid     VARCHAR(255) UNIQUE NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction   VARCHAR(255) NOT NULL,
    request_id    VARCHAR(255),
    currency      CHAR(3)      NOT NULL,
    provider      VARCHAR(255) NOT NULL,
    amount        INT          NOT NULL,
    payment_dt    BIGINT       NOT NULL,
    bank          VARCHAR(255) NOT NULL,
    delivery_cost INT          NOT NULL,
    goods_total   INT          NOT NULL,
    custom_fee    INT          NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
    id           SERIAL PRIMARY KEY,
    order_uid    VARCHAR(255) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id      INT          NOT NULL,
    track_number VARCHAR(255) NOT NULL,
    price        INT          NOT NULL,
    rid          VARCHAR(255) NOT NULL,
    name         VARCHAR(255) NOT NULL,
    sale         INT          NOT NULL DEFAULT 0,
    size         VARCHAR(255) NOT NULL,
    total_price  INT          NOT NULL,
    nm_id        INT          NOT NULL,
    brand        VARCHAR(255) NOT NULL,
    status       INT          NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders(date_created);
CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid);

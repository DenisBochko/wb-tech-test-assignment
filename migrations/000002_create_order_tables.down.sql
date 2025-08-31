-- 000002_create_order_tables.down.sql

DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS items;

DROP INDEX IF EXISTS idx_orders_customer;
DROP INDEX IF EXISTS idx_orders_date_created;
DROP INDEX IF EXISTS idx_items_order_uid;

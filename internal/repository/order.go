package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wb-tech-test-assignment/internal/apperrors"
	"wb-tech-test-assignment/internal/model"
)

// Пояснение: Можно было бы сделать 2 запроса SELECT в бд, но т.к. полей в таблицах довольно много
// решил разнести это по отдельным запросам.

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (o *OrderRepository) PutOrder(ctx context.Context, order model.Order) error {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = o.insertOrder(ctx, tx, order)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	err = o.insertDelivery(ctx, tx, order.OrderUID, order.Delivery)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	err = o.insertPayment(ctx, tx, order.OrderUID, order.Payment)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	err = o.insertItems(ctx, tx, order.OrderUID, order.Items)
	if err != nil {
		return fmt.Errorf("failed to insert items: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (o *OrderRepository) GetOrder(ctx context.Context, orderUID string) (model.Order, error) {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	order, err := o.selectOrder(ctx, tx, orderUID)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to select order: %w", err)
	}

	delivery, err := o.selectDelivery(ctx, tx, orderUID)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to select delivery: %w", err)
	}

	payment, err := o.selectPayment(ctx, tx, orderUID)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to select payment: %w", err)
	}

	items, err := o.selectItems(ctx, tx, orderUID)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to select items: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return model.Order{}, fmt.Errorf("error committing transaction: %w", err)
	}

	order.Delivery = delivery
	order.Payment = payment
	order.Items = items

	return order, nil
}

func (o *OrderRepository) insertOrder(ctx context.Context, ext RepoExtension, order model.Order) error {
	if ext == nil {
		ext = o.db
	}

	const query = `
		INSERT INTO orders (order_uid, 
		                    track_number, 
		                    entry, 
		                    locale, 
		                    internal_signature, 
		                    customer_id, 
		                    delivery_service, 
		                    shardkey, 
		                    sm_id, 
		                    date_created, 
		                    oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`

	_, err := ext.Exec(ctx, query,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.ShardKey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		return err
	}

	return nil
}

func (o *OrderRepository) selectOrder(ctx context.Context, ext RepoExtension, OrderUID string) (model.Order, error) {
	if ext == nil {
		ext = o.db
	}

	const query = `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1;
	`

	var order model.Order

	err := ext.QueryRow(ctx, query, OrderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, apperrors.ErrOrderNotFound
		}

		return model.Order{}, err
	}

	return order, nil
}

func (o *OrderRepository) insertDelivery(ctx context.Context, ext RepoExtension, orderUID string, delivery model.Delivery) error {
	if ext == nil {
		ext = o.db
	}

	const query = `
		INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`

	_, err := ext.Exec(ctx, query,
		orderUID,
		delivery.Name,
		delivery.Phone,
		delivery.Zip,
		delivery.City,
		delivery.Address,
		delivery.Region,
		delivery.Email,
	)
	if err != nil {
		return err
	}

	return nil
}

func (o *OrderRepository) selectDelivery(ctx context.Context, ext RepoExtension, orderUID string) (model.Delivery, error) {
	if ext == nil {
		ext = o.db
	}

	const query = `
		SELECT name, phone, zip, city, address, region, email
		FROM deliveries
		WHERE order_uid = $1;
	`

	var delivery model.Delivery

	err := ext.QueryRow(ctx, query, orderUID).Scan(
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)
	if err != nil {
		return model.Delivery{}, err
	}

	return delivery, nil
}

func (o *OrderRepository) insertPayment(ctx context.Context, ext RepoExtension, orderUID string, payment model.Payment) error {
	if ext == nil {
		ext = o.db
	}

	const query = `
		INSERT INTO payments (order_uid, 
		                      transaction, 
		                      request_id, 
		                      currency, 
		                      provider, 
		                      amount, 
		                      payment_dt, 
		                      bank, 
		                      delivery_cost, 
		                      goods_total,
		                      custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`

	_, err := ext.Exec(ctx, query,
		orderUID,
		payment.Transaction,
		payment.RequestID,
		payment.Currency,
		payment.Provider,
		payment.Amount,
		payment.PaymentDt,
		payment.Bank,
		payment.DeliveryCost,
		payment.GoodsTotal,
		payment.CustomFee,
	)
	if err != nil {
		return err
	}

	return nil
}

func (o *OrderRepository) selectPayment(ctx context.Context, ext RepoExtension, orderUID string) (model.Payment, error) {
	if ext == nil {
		ext = o.db
	}

	const query = `
		SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments
		WHERE order_uid = $1;
	`

	var payment model.Payment

	err := ext.QueryRow(ctx, query, orderUID).Scan(
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDt,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
	)
	if err != nil {
		return model.Payment{}, err
	}

	return payment, nil
}

func (o *OrderRepository) insertItems(ctx context.Context, ext RepoExtension, orderUID string, items []model.Item) error {
	if ext == nil {
		ext = o.db
	}

	const query = `
		INSERT INTO items (order_uid, 
		                   chrt_id, 
		                   track_number, 
		                   price, 
		                   rid, 
		                   name, 
		                   sale, 
		                   size, 
		                   total_price, 
		                   nm_id, 
		                   brand, 
		                   status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
    `

	batch := &pgx.Batch{}
	for _, v := range items {
		batch.Queue(query,
			orderUID,
			v.ChrtID,
			v.TrackNumber,
			v.Price,
			v.RID,
			v.Name,
			v.Sale,
			v.Size,
			v.TotalPrice,
			v.NmID,
			v.Brand,
			v.Status,
		)
	}

	br := ext.SendBatch(ctx, batch)
	defer func() {
		_ = br.Close()
	}()

	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (o *OrderRepository) selectItems(ctx context.Context, ext RepoExtension, orderUID string) ([]model.Item, error) {
	if ext == nil {
		ext = o.db
	}

	const query = `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1;
	`

	var items []model.Item

	rows, err := ext.Query(ctx, query, orderUID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item model.Item

		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.RID,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

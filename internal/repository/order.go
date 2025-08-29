package repository

import (
	"context"
	"errors"
	"wb-tech-test-assignment/internal/apperrors"
	"wb-tech-test-assignment/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Пояснение: Можно было бы сделать 2 запроса SELECT в бд, но т.к. полей в таблицах довольно много
// решил разнести это по отдельным запросам.

type OrderRepository struct {
	DB *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		DB: db,
	}
}

func (o *OrderRepository) InsertOrder(ctx context.Context, ext RepoExtension, order model.Order) error {
	if ext == nil {
		ext = o.DB
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

func (o *OrderRepository) SelectOrder(ctx context.Context, ext RepoExtension, OrderUID string) (model.Order, error) {
	if ext == nil {
		ext = o.DB
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

func (o *OrderRepository) InsertDelivery(ctx context.Context, ext RepoExtension, orderUID string, delivery model.Delivery) error {
	if ext == nil {
		ext = o.DB
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

func (o *OrderRepository) SelectDelivery(ctx context.Context, ext RepoExtension, orderUID string) (model.Delivery, error) {
	if ext == nil {
		ext = o.DB
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

func (o *OrderRepository) InsertPayment(ctx context.Context, ext RepoExtension, orderUID string, payment model.Payment) error {
	if ext == nil {
		ext = o.DB
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

func (o *OrderRepository) SelectPayment(ctx context.Context, ext RepoExtension, orderUID string) (model.Payment, error) {
	if ext == nil {
		ext = o.DB
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

func (o *OrderRepository) InsertItems(ctx context.Context, ext RepoExtension, orderUID string, items []model.Item) error {
	if ext == nil {
		ext = o.DB
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

func (o *OrderRepository) SelectItems(ctx context.Context, ext RepoExtension, orderUID string) ([]model.Item, error) {
	if ext == nil {
		ext = o.DB
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

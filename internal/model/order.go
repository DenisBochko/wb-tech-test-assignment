package model

import (
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid"         validate:"required"`
	TrackNumber       string    `json:"track_number"      validate:"required"`
	Entry             string    `json:"entry"             validate:"required"`
	Delivery          Delivery  `json:"delivery"          validate:"required,dive"`
	Payment           Payment   `json:"payment"           validate:"required,dive"`
	Items             []Item    `json:"items"             validate:"required,dive,min=1"`
	Locale            string    `json:"locale"            validate:"required,len=2"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"       validate:"required"`
	DeliveryService   string    `json:"delivery_service"  validate:"required"`
	ShardKey          string    `json:"shardkey"          validate:"required,numeric"`
	SmID              int       `json:"sm_id"             validate:"required"`
	DateCreated       time.Time `json:"date_created"      validate:"required"`
	OofShard          string    `json:"oof_shard"         validate:"required,numeric"`
}

type Delivery struct {
	Name    string `json:"name"    validate:"required,min=1,max=255"`
	Phone   string `json:"phone"   validate:"required"`
	Zip     string `json:"zip"     validate:"required,numeric"`
	City    string `json:"city"    validate:"required,min=1,max=255"`
	Address string `json:"address" validate:"required,min=1,max=255"`
	Region  string `json:"region"  validate:"required,min=1,max=255"`
	Email   string `json:"email"   validate:"required,email"`
}

type Payment struct {
	Transaction  string `json:"transaction"   validate:"required"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"      validate:"required,len=3,uppercase"` // ISO 4217
	Provider     string `json:"provider"      validate:"required"`
	Amount       int    `json:"amount"        validate:"required,gt=0"`
	PaymentDt    int64  `json:"payment_dt"    validate:"required,gt=0"`
	Bank         string `json:"bank"          validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"required,gte=0"`
	GoodsTotal   int    `json:"goods_total"   validate:"required,gte=0"`
	CustomFee    int    `json:"custom_fee"    validate:"gte=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"      validate:"required,gt=0"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price"        validate:"required,gt=0"`
	RID         string `json:"rid"          validate:"required"`
	Name        string `json:"name"         validate:"required"`
	Sale        int    `json:"sale"         validate:"gte=0"`
	Size        string `json:"size"         validate:"required"`
	TotalPrice  int    `json:"total_price"  validate:"required,gte=0"`
	NmID        int    `json:"nm_id"        validate:"required,gt=0"`
	Brand       string `json:"brand"        validate:"required"`
	Status      int    `json:"status"       validate:"required,gt=0"`
}

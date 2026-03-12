package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ProductID primitive.ObjectID `json:"product_id" bson:"product_id"`
	Name      string             `json:"name" bson:"name"`
	Quantity  int                `json:"quantity" bson:"quantity"`
	UnitPrice float64            `json:"unit_price" bson:"unit_price"`
}

type ShippingAddress struct {
	Name       string `json:"name" bson:"name"`
	Street     string `json:"street" bson:"street"`
	Number     string `json:"number" bson:"number"`
	Complement string `json:"complement,omitempty" bson:"complement,omitempty"`
	District   string `json:"district" bson:"district"`
	City       string `json:"city" bson:"city"`
	State      string `json:"state" bson:"state"`
	ZipCode    string `json:"zip_code" bson:"zip_code"`
	Phone      string `json:"phone,omitempty" bson:"phone,omitempty"`
}

type Order struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID          primitive.ObjectID `json:"user_id" bson:"user_id"`
	Items           []OrderItem        `json:"items" bson:"items"`
	Total           float64            `json:"total" bson:"total"`
	Status          string             `json:"status" bson:"status"`
	ShippingAddress ShippingAddress    `json:"shipping_address" bson:"shipping_address"`
	PaymentStatus   string             `json:"payment_status" bson:"payment_status"`
	Notes           string             `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// Status values: pending, confirmed, printing, shipped, delivered, cancelled
// PaymentStatus values: pending, paid, refunded, failed

type CreateOrderItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CreateOrderRequest struct {
	Items           []CreateOrderItemRequest `json:"items"`
	ShippingAddress ShippingAddress          `json:"shipping_address"`
	Notes           string                   `json:"notes,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status        string `json:"status,omitempty"`
	PaymentStatus string `json:"payment_status,omitempty"`
}

type OrderListResponse struct {
	Orders []Order `json:"orders"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}

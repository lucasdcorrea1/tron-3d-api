package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name"`
	Slug           string             `json:"slug" bson:"slug"`
	Description    string             `json:"description,omitempty" bson:"description,omitempty"`
	Price          float64            `json:"price" bson:"price"`
	Images         []string           `json:"images" bson:"images"`
	CategoryID     primitive.ObjectID `json:"category_id" bson:"category_id"`
	Material       string             `json:"material,omitempty" bson:"material,omitempty"`
	Dimensions     string             `json:"dimensions,omitempty" bson:"dimensions,omitempty"`
	Weight         float64            `json:"weight,omitempty" bson:"weight,omitempty"`
	PrintTimeHours float64            `json:"print_time_hours,omitempty" bson:"print_time_hours,omitempty"`
	Stock          int                `json:"stock" bson:"stock"`
	Active         bool               `json:"active" bson:"active"`
	Featured       bool               `json:"featured" bson:"featured"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

type CreateProductRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Price          float64  `json:"price"`
	Images         []string `json:"images,omitempty"`
	CategoryID     string   `json:"category_id"`
	Material       string   `json:"material,omitempty"`
	Dimensions     string   `json:"dimensions,omitempty"`
	Weight         float64  `json:"weight,omitempty"`
	PrintTimeHours float64  `json:"print_time_hours,omitempty"`
	Stock          int      `json:"stock"`
	Featured       bool     `json:"featured"`
}

type UpdateProductRequest struct {
	Name           *string  `json:"name,omitempty"`
	Description    *string  `json:"description,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	Images         []string `json:"images,omitempty"`
	CategoryID     *string  `json:"category_id,omitempty"`
	Material       *string  `json:"material,omitempty"`
	Dimensions     *string  `json:"dimensions,omitempty"`
	Weight         *float64 `json:"weight,omitempty"`
	PrintTimeHours *float64 `json:"print_time_hours,omitempty"`
	Stock          *int     `json:"stock,omitempty"`
	Active         *bool    `json:"active,omitempty"`
	Featured       *bool    `json:"featured,omitempty"`
}

type ProductListResponse struct {
	Products []Product `json:"products"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

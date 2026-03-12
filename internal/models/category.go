package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID          primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Name        string              `json:"name" bson:"name"`
	Slug        string              `json:"slug" bson:"slug"`
	Description string              `json:"description,omitempty" bson:"description,omitempty"`
	Image       string              `json:"image,omitempty" bson:"image,omitempty"`
	ParentID    *primitive.ObjectID `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	SortOrder   int                 `json:"sort_order" bson:"sort_order"`
	Active      bool                `json:"active" bson:"active"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" bson:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Image       *string `json:"image,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

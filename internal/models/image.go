package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Image struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UploaderID primitive.ObjectID `json:"uploader_id" bson:"uploader_id"`
	GroupID    string             `json:"group_id,omitempty" bson:"group_id,omitempty"`
	SizeLabel  string             `json:"size_label,omitempty" bson:"size_label,omitempty"`
	Width      int                `json:"width,omitempty" bson:"width,omitempty"`
	Data       string             `json:"-" bson:"data"`
	Size       int                `json:"size" bson:"size"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
}

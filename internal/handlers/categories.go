package handlers

import (
	"net/http"

	"github.com/tron-legacy/tron-3d-api/internal/database"
	"github.com/tron-legacy/tron-3d-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ListCategories returns all active categories sorted by sort_order
func ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := bson.M{"active": true}
	opts := options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}, {Key: "name", Value: 1}})

	cursor, err := database.Categories().Find(ctx, filter, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to decode categories")
		return
	}

	if categories == nil {
		categories = []models.Category{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"categories": categories})
}

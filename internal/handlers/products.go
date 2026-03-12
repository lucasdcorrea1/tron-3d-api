package handlers

import (
	"net/http"

	"github.com/tron-legacy/tron-3d-api/internal/database"
	"github.com/tron-legacy/tron-3d-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ListProducts returns active products with pagination and filters
func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 12)
	if limit > 50 {
		limit = 50
	}

	filter := bson.M{"active": true}

	// Filter by category_id
	if catID := r.URL.Query().Get("category_id"); catID != "" {
		objID, err := primitive.ObjectIDFromHex(catID)
		if err == nil {
			filter["category_id"] = objID
		}
	}

	// Filter by featured
	if featured := r.URL.Query().Get("featured"); featured == "true" {
		filter["featured"] = true
	}

	// Text search
	if search := r.URL.Query().Get("search"); search != "" {
		filter["$text"] = bson.M{"$search": search}
	}

	// Count total
	total, err := database.Products().CountDocuments(ctx, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count products")
		return
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "featured", Value: -1}, {Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := database.Products().Find(ctx, filter, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to decode products")
		return
	}

	if products == nil {
		products = []models.Product{}
	}

	writeJSON(w, http.StatusOK, models.ProductListResponse{
		Products: products,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

// GetProductBySlug returns a single active product by slug
func GetProductBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := r.PathValue("slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "Slug is required")
		return
	}

	var product models.Product
	err := database.Products().FindOne(ctx, bson.M{"slug": slug, "active": true}).Decode(&product)
	if err != nil {
		writeError(w, http.StatusNotFound, "Product not found")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

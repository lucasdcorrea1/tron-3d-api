package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tron-legacy/tron-3d-api/internal/database"
	"github.com/tron-legacy/tron-3d-api/internal/middleware"
	"github.com/tron-legacy/tron-3d-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AdminListProducts returns all products (including inactive)
func AdminListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	filter := bson.M{}

	if catID := r.URL.Query().Get("category_id"); catID != "" {
		objID, err := primitive.ObjectIDFromHex(catID)
		if err == nil {
			filter["category_id"] = objID
		}
	}

	if active := r.URL.Query().Get("active"); active != "" {
		filter["active"] = active == "true"
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filter["$text"] = bson.M{"$search": search}
	}

	total, err := database.Products().CountDocuments(ctx, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count products")
		return
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
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

// AdminCreateProduct creates a new product
func AdminCreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if req.Price <= 0 {
		writeError(w, http.StatusBadRequest, "Price must be positive")
		return
	}

	categoryID, err := primitive.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid category_id")
		return
	}

	// Generate unique slug
	baseSlug := slugify(req.Name)
	slug := baseSlug
	counter := 1
	for {
		count, _ := database.Products().CountDocuments(ctx, bson.M{"slug": slug})
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	if req.Images == nil {
		req.Images = []string{}
	}

	now := time.Now()
	product := models.Product{
		Name:           req.Name,
		Slug:           slug,
		Description:    req.Description,
		Price:          req.Price,
		Images:         req.Images,
		CategoryID:     categoryID,
		Material:       req.Material,
		Dimensions:     req.Dimensions,
		Weight:         req.Weight,
		PrintTimeHours: req.PrintTimeHours,
		Stock:          req.Stock,
		Active:         true,
		Featured:       req.Featured,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result, err := database.Products().InsertOne(ctx, product)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	product.ID = result.InsertedID.(primitive.ObjectID)
	middleware.IncProductCreated()

	writeJSON(w, http.StatusCreated, product)
}

// AdminUpdateProduct updates an existing product
func AdminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productIDStr := r.PathValue("id")
	productID, err := primitive.ObjectIDFromHex(productIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	update := bson.M{"updated_at": time.Now()}

	if req.Name != nil {
		update["name"] = *req.Name
		// Regenerate slug if name changed
		baseSlug := slugify(*req.Name)
		slug := baseSlug
		counter := 1
		for {
			count, _ := database.Products().CountDocuments(ctx, bson.M{"slug": slug, "_id": bson.M{"$ne": productID}})
			if count == 0 {
				break
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, counter)
			counter++
		}
		update["slug"] = slug
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Price != nil {
		update["price"] = *req.Price
	}
	if req.Images != nil {
		update["images"] = req.Images
	}
	if req.CategoryID != nil {
		catID, err := primitive.ObjectIDFromHex(*req.CategoryID)
		if err == nil {
			update["category_id"] = catID
		}
	}
	if req.Material != nil {
		update["material"] = *req.Material
	}
	if req.Dimensions != nil {
		update["dimensions"] = *req.Dimensions
	}
	if req.Weight != nil {
		update["weight"] = *req.Weight
	}
	if req.PrintTimeHours != nil {
		update["print_time_hours"] = *req.PrintTimeHours
	}
	if req.Stock != nil {
		update["stock"] = *req.Stock
	}
	if req.Active != nil {
		update["active"] = *req.Active
	}
	if req.Featured != nil {
		update["featured"] = *req.Featured
	}

	result, err := database.Products().UpdateOne(ctx, bson.M{"_id": productID}, bson.M{"$set": update})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	if result.MatchedCount == 0 {
		writeError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Return updated product
	var product models.Product
	database.Products().FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	writeJSON(w, http.StatusOK, product)
}

// AdminDeleteProduct deletes a product
func AdminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productIDStr := r.PathValue("id")
	productID, err := primitive.ObjectIDFromHex(productIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	result, err := database.Products().DeleteOne(ctx, bson.M{"_id": productID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	if result.DeletedCount == 0 {
		writeError(w, http.StatusNotFound, "Product not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Product deleted"})
}

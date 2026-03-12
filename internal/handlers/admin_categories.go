package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tron-legacy/tron-3d-api/internal/database"
	"github.com/tron-legacy/tron-3d-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminCreateCategory creates a new category
func AdminCreateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Generate unique slug
	baseSlug := slugify(req.Name)
	slug := baseSlug
	counter := 1
	for {
		count, _ := database.Categories().CountDocuments(ctx, bson.M{"slug": slug})
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	var parentID *primitive.ObjectID
	if req.ParentID != "" {
		pid, err := primitive.ObjectIDFromHex(req.ParentID)
		if err == nil {
			parentID = &pid
		}
	}

	now := time.Now()
	category := models.Category{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Image:       req.Image,
		ParentID:    parentID,
		SortOrder:   req.SortOrder,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := database.Categories().InsertOne(ctx, category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}

	category.ID = result.InsertedID.(primitive.ObjectID)
	writeJSON(w, http.StatusCreated, category)
}

// AdminUpdateCategory updates an existing category
func AdminUpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	catIDStr := r.PathValue("id")
	catID, err := primitive.ObjectIDFromHex(catIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	update := bson.M{"updated_at": time.Now()}

	if req.Name != nil {
		update["name"] = *req.Name
		baseSlug := slugify(*req.Name)
		slug := baseSlug
		counter := 1
		for {
			count, _ := database.Categories().CountDocuments(ctx, bson.M{"slug": slug, "_id": bson.M{"$ne": catID}})
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
	if req.Image != nil {
		update["image"] = *req.Image
	}
	if req.ParentID != nil {
		if *req.ParentID == "" {
			update["parent_id"] = nil
		} else {
			pid, err := primitive.ObjectIDFromHex(*req.ParentID)
			if err == nil {
				update["parent_id"] = pid
			}
		}
	}
	if req.SortOrder != nil {
		update["sort_order"] = *req.SortOrder
	}
	if req.Active != nil {
		update["active"] = *req.Active
	}

	result, err := database.Categories().UpdateOne(ctx, bson.M{"_id": catID}, bson.M{"$set": update})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update category")
		return
	}

	if result.MatchedCount == 0 {
		writeError(w, http.StatusNotFound, "Category not found")
		return
	}

	var category models.Category
	database.Categories().FindOne(ctx, bson.M{"_id": catID}).Decode(&category)
	writeJSON(w, http.StatusOK, category)
}

// AdminDeleteCategory deletes a category
func AdminDeleteCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	catIDStr := r.PathValue("id")
	catID, err := primitive.ObjectIDFromHex(catIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// Check if category has products
	count, _ := database.Products().CountDocuments(ctx, bson.M{"category_id": catID})
	if count > 0 {
		writeError(w, http.StatusConflict, "Cannot delete category with products")
		return
	}

	result, err := database.Categories().DeleteOne(ctx, bson.M{"_id": catID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	if result.DeletedCount == 0 {
		writeError(w, http.StatusNotFound, "Category not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Category deleted"})
}

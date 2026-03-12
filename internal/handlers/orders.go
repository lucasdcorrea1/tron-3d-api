package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tron-legacy/tron-3d-api/internal/database"
	"github.com/tron-legacy/tron-3d-api/internal/middleware"
	"github.com/tron-legacy/tron-3d-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateOrder creates a new order (auth required)
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(r)
	if userID.IsZero() {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "Order must have at least one item")
		return
	}

	// Validate shipping address
	if req.ShippingAddress.Name == "" || req.ShippingAddress.Street == "" ||
		req.ShippingAddress.City == "" || req.ShippingAddress.State == "" ||
		req.ShippingAddress.ZipCode == "" {
		writeError(w, http.StatusBadRequest, "Shipping address is incomplete")
		return
	}

	// Resolve items: validate products, check stock, calculate total
	var orderItems []models.OrderItem
	var total float64

	for _, item := range req.Items {
		if item.Quantity < 1 {
			writeError(w, http.StatusBadRequest, "Quantity must be at least 1")
			return
		}

		productID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid product ID: "+item.ProductID)
			return
		}

		var product models.Product
		err = database.Products().FindOne(ctx, bson.M{"_id": productID, "active": true}).Decode(&product)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Product not found: "+item.ProductID)
			return
		}

		if product.Stock < item.Quantity {
			writeError(w, http.StatusBadRequest, "Insufficient stock for: "+product.Name)
			return
		}

		orderItems = append(orderItems, models.OrderItem{
			ProductID: productID,
			Name:      product.Name,
			Quantity:  item.Quantity,
			UnitPrice: product.Price,
		})

		total += product.Price * float64(item.Quantity)
	}

	// Decrement stock
	for _, item := range orderItems {
		_, err := database.Products().UpdateOne(ctx,
			bson.M{"_id": item.ProductID},
			bson.M{"$inc": bson.M{"stock": -item.Quantity}},
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update stock")
			return
		}
	}

	now := time.Now()
	order := models.Order{
		UserID:          userID,
		Items:           orderItems,
		Total:           total,
		Status:          "pending",
		ShippingAddress: req.ShippingAddress,
		PaymentStatus:   "pending",
		Notes:           req.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result, err := database.Orders().InsertOne(ctx, order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}

	order.ID = result.InsertedID.(primitive.ObjectID)
	middleware.IncOrderCreated()

	writeJSON(w, http.StatusCreated, order)
}

// ListMyOrders returns orders for the authenticated user
func ListMyOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(r)
	if userID.IsZero() {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	filter := bson.M{"user_id": userID}

	total, err := database.Orders().CountDocuments(ctx, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count orders")
		return
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := database.Orders().Find(ctx, filter, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to decode orders")
		return
	}

	if orders == nil {
		orders = []models.Order{}
	}

	writeJSON(w, http.StatusOK, models.OrderListResponse{
		Orders: orders,
		Total:  total,
		Page:   page,
		Limit:  limit,
	})
}

// GetOrder returns a single order (verifies ownership)
func GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(r)
	if userID.IsZero() {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orderIDStr := r.PathValue("id")
	orderID, err := primitive.ObjectIDFromHex(orderIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var order models.Order
	err = database.Orders().FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		writeError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Verify ownership
	if order.UserID != userID {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	writeJSON(w, http.StatusOK, order)
}

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tron-legacy/tron-3d-api/internal/database"
	"github.com/tron-legacy/tron-3d-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AdminListOrders returns all orders with optional status filter
func AdminListOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	filter := bson.M{}

	if status := r.URL.Query().Get("status"); status != "" {
		filter["status"] = status
	}

	if paymentStatus := r.URL.Query().Get("payment_status"); paymentStatus != "" {
		filter["payment_status"] = paymentStatus
	}

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

// AdminUpdateOrderStatus updates the status of an order
func AdminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderIDStr := r.PathValue("id")
	orderID, err := primitive.ObjectIDFromHex(orderIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	update := bson.M{"updated_at": time.Now()}

	validStatuses := map[string]bool{
		"pending": true, "confirmed": true, "printing": true,
		"shipped": true, "delivered": true, "cancelled": true,
	}
	validPaymentStatuses := map[string]bool{
		"pending": true, "paid": true, "refunded": true, "failed": true,
	}

	if req.Status != "" {
		if !validStatuses[req.Status] {
			writeError(w, http.StatusBadRequest, "Invalid status")
			return
		}
		update["status"] = req.Status
	}

	if req.PaymentStatus != "" {
		if !validPaymentStatuses[req.PaymentStatus] {
			writeError(w, http.StatusBadRequest, "Invalid payment status")
			return
		}
		update["payment_status"] = req.PaymentStatus
	}

	result, err := database.Orders().UpdateOne(ctx, bson.M{"_id": orderID}, bson.M{"$set": update})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update order")
		return
	}

	if result.MatchedCount == 0 {
		writeError(w, http.StatusNotFound, "Order not found")
		return
	}

	// If cancelled, restore stock
	if req.Status == "cancelled" {
		var order models.Order
		database.Orders().FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
		for _, item := range order.Items {
			database.Products().UpdateOne(ctx,
				bson.M{"_id": item.ProductID},
				bson.M{"$inc": bson.M{"stock": item.Quantity}},
			)
		}
	}

	var order models.Order
	database.Orders().FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	writeJSON(w, http.StatusOK, order)
}

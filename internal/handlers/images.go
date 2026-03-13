package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"time"

	"github.com/tron-legacy/tron-3d-api/internal/database"
	"github.com/tron-legacy/tron-3d-api/internal/middleware"
	"github.com/tron-legacy/tron-3d-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	_ "golang.org/x/image/webp"
	"golang.org/x/image/draw"
)

type sizeVariant struct {
	Label   string
	Width   int
	Quality int
}

var variants = []sizeVariant{
	{"thumb", 400, 55},
	{"card", 800, 65},
	{"full", 1200, 75},
}

// UploadImage handles multipart image upload and creates 3 size variants.
// POST /api/v1/admin/upload (admin only)
func UploadImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(r)

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20) // 5 MB

	file, _, err := r.FormFile("image")
	if err != nil {
		if err.Error() == "http: request body too large" {
			writeError(w, http.StatusRequestEntityTooLarge, "File too large (max 5MB)")
			return
		}
		writeError(w, http.StatusBadRequest, "Missing or invalid image field")
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "File too large (max 5MB)")
		return
	}

	ct := http.DetectContentType(raw)
	if ct != "image/jpeg" && ct != "image/png" && ct != "image/webp" {
		writeError(w, http.StatusBadRequest, "Only JPEG, PNG, and WebP images are accepted")
		return
	}

	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Could not decode image")
		return
	}

	groupID := primitive.NewObjectID().Hex()
	now := time.Now()

	type variantResult struct {
		Label string `json:"label"`
		ID    string `json:"id"`
		URL   string `json:"url"`
	}
	results := make(map[string]string)

	for _, v := range variants {
		resized := resizeImage(src, v.Width)

		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: v.Quality}); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to encode variant: "+v.Label)
			return
		}

		b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

		doc := models.Image{
			UploaderID: userID,
			GroupID:    groupID,
			SizeLabel:  v.Label,
			Width:      resized.Bounds().Dx(),
			Data:       b64,
			Size:       buf.Len(),
			CreatedAt:  now,
		}

		res, err := database.Images().InsertOne(ctx, doc)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save variant: "+v.Label)
			return
		}

		oid := res.InsertedID.(primitive.ObjectID)
		results[v.Label] = fmt.Sprintf("/api/v1/images/%s", oid.Hex())
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"group_id": groupID,
		"thumb":    results["thumb"],
		"card":     results["card"],
		"full":     results["full"],
	})
}

// ServeImage serves a single image by its _id.
// GET /api/v1/images/{id} (public)
func ServeImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid image ID")
		return
	}

	var img models.Image
	err = database.Images().FindOne(ctx, bson.M{"_id": oid}).Decode(&img)
	if err != nil {
		writeError(w, http.StatusNotFound, "Image not found")
		return
	}

	data, err := base64.StdEncoding.DecodeString(img.Data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Corrupt image data")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data)
}

// ServeImageByGroup serves an image variant by group_id and optional size query param.
// GET /api/v1/images/group/{groupId}?size=thumb|card|full (public)
func ServeImageByGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	groupID := r.PathValue("groupId")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "Group ID is required")
		return
	}

	size := r.URL.Query().Get("size")
	if size == "" {
		size = "card"
	}
	if size != "thumb" && size != "card" && size != "full" {
		writeError(w, http.StatusBadRequest, "Invalid size. Use: thumb, card, or full")
		return
	}

	var img models.Image
	err := database.Images().FindOne(ctx, bson.M{
		"group_id":   groupID,
		"size_label": size,
	}).Decode(&img)
	if err != nil {
		writeError(w, http.StatusNotFound, "Image not found")
		return
	}

	data, err := base64.StdEncoding.DecodeString(img.Data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Corrupt image data")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data)
}

// DeleteImageGroup deletes all image variants for a group_id.
// DELETE /api/v1/admin/images/{groupId} (admin only)
func DeleteImageGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	groupID := r.PathValue("groupId")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "Group ID is required")
		return
	}

	result, err := database.Images().DeleteMany(ctx, bson.M{"group_id": groupID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete images")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": result.DeletedCount,
	})
}

// resizeImage scales src to the given maxWidth preserving aspect ratio using CatmullRom.
// If the source is already smaller than maxWidth, it is returned as-is.
func resizeImage(src image.Image, maxWidth int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW <= maxWidth {
		return src
	}

	ratio := float64(maxWidth) / float64(srcW)
	newW := maxWidth
	newH := int(float64(srcH) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}

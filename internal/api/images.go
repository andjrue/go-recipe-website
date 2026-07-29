package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"recipe-website/internal/repository"
	"recipe-website/internal/storage"
)

const (
	maxImageRequestBytes = 20 << 20
	maxImagePixels       = 12_000_000
)

type ImageHandler struct {
	recipes repository.RecipeRepository
	store   storage.ImageStore
}

func NewImageHandler(recipes repository.RecipeRepository, store storage.ImageStore) *ImageHandler {
	return &ImageHandler{recipes: recipes, store: store}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	recipeID, ok := recipeIDFromRequest(w, r)
	if !ok {
		return
	}
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "image_storage_unavailable")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxImageRequestBytes)
	file, header, err := r.FormFile("image")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "image_too_large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_image")
		return
	}
	defer file.Close()

	normalized, err := normalizeJPEG(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_image")
		return
	}
	key, err := imageStorageKey()
	if err != nil {
		log.Printf("generating image key: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	if err := h.store.Save(r.Context(), key, bytes.NewReader(normalized)); err != nil {
		log.Printf("storing recipe image: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	imageRecord, err := h.recipes.AddImage(r.Context(), recipeID, &repository.RecipeImage{
		S3Key: key, FileName: safeImageName(header), ContentType: "image/jpeg", FileSize: int64(len(normalized)),
	})
	if err != nil {
		if cleanupErr := h.store.Delete(r.Context(), key); cleanupErr != nil {
			log.Printf("cleaning up image after database failure: %v", cleanupErr)
		}
		writeRepositoryError(w, "adding recipe image", err)
		return
	}

	writeJSON(w, http.StatusCreated, newImageResponse(imageRecord))
}

func (h *ImageHandler) Serve(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "image_storage_unavailable")
		return
	}
	imageID := r.PathValue("id")
	if !isUUID(imageID) {
		writeError(w, http.StatusBadRequest, "invalid_id")
		return
	}
	imageRecord, err := h.recipes.GetImage(r.Context(), imageID)
	if err != nil {
		writeRepositoryError(w, "getting recipe image", err)
		return
	}
	file, err := h.store.Open(r.Context(), imageRecord.S3Key)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		log.Printf("opening recipe image: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("serving recipe image: %v", err)
	}
}

func (h *ImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "image_storage_unavailable")
		return
	}
	recipeID := r.PathValue("recipeID")
	imageID := r.PathValue("imageID")
	if !isUUID(recipeID) || !isUUID(imageID) {
		writeError(w, http.StatusBadRequest, "invalid_id")
		return
	}
	imageRecord, err := h.recipes.DeleteImage(r.Context(), recipeID, imageID)
	if err != nil {
		writeRepositoryError(w, "deleting recipe image", err)
		return
	}
	if err := h.store.Delete(r.Context(), imageRecord.S3Key); err != nil {
		log.Printf("deleting stored recipe image %s: %v", imageRecord.ID, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ImageHandler) SetCover(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("recipeID")
	imageID := r.PathValue("imageID")
	if !isUUID(recipeID) || !isUUID(imageID) {
		writeError(w, http.StatusBadRequest, "invalid_id")
		return
	}
	if err := h.recipes.SetCoverImage(r.Context(), recipeID, imageID); err != nil {
		writeRepositoryError(w, "setting recipe cover", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func normalizeJPEG(source multipart.File) ([]byte, error) {
	data, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("reading image: %w", err)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > maxImagePixels {
		return nil, errors.New("unsupported image")
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}
	var normalized bytes.Buffer
	if err := jpeg.Encode(&normalized, decoded, &jpeg.Options{Quality: 88}); err != nil {
		return nil, fmt.Errorf("encoding image: %w", err)
	}
	return normalized.Bytes(), nil
}

func imageStorageKey() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value) + ".jpg", nil
}

func safeImageName(header *multipart.FileHeader) string {
	name := strings.TrimSpace(filepath.Base(header.Filename))
	if name == "" || name == "." {
		return "recipe-photo.jpg"
	}
	name = strings.TrimSuffix(name, filepath.Ext(name)) + ".jpg"
	if len(name) > 200 {
		name = name[:196] + ".jpg"
	}
	return name
}

func newImageResponse(image *repository.RecipeImage) imageResponse {
	return imageResponse{
		ID: image.ID, FileName: image.FileName, ContentType: image.ContentType, FileSize: image.FileSize,
		Position: image.Position, IsCover: image.IsCover, UploadedAt: image.UploadedAt,
		URL: "/api/recipe-images/" + image.ID,
	}
}

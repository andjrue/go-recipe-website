package api

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"recipe-website/internal/repository"
	"recipe-website/internal/storage"
)

const testImageID = "55555555-5555-4555-8555-555555555555"

type memoryImageStore struct {
	data map[string][]byte
}

func (s *memoryImageStore) Save(_ context.Context, key string, source io.Reader) error {
	data, err := io.ReadAll(source)
	if err != nil {
		return err
	}
	if s.data == nil {
		s.data = make(map[string][]byte)
	}
	s.data[key] = data
	return nil
}

func (s *memoryImageStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	data, ok := s.data[key]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *memoryImageStore) Delete(_ context.Context, key string) error {
	delete(s.data, key)
	return nil
}

func TestUploadNormalizesAndStoresImage(t *testing.T) {
	var source bytes.Buffer
	imageSource := image.NewRGBA(image.Rect(0, 0, 4, 3))
	imageSource.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&source, imageSource); err != nil {
		t.Fatalf("encoding source image: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "card.png")
	if err != nil {
		t.Fatalf("creating multipart image: %v", err)
	}
	if _, err := part.Write(source.Bytes()); err != nil {
		t.Fatalf("writing multipart image: %v", err)
	}
	writer.Close()

	store := &memoryImageStore{}
	recipes := fakeRecipeRepository{
		addImage: func(_ context.Context, recipeID string, image *repository.RecipeImage) (*repository.RecipeImage, error) {
			if recipeID != testRecipeID || image.ContentType != "image/jpeg" || image.FileName != "card.jpg" || image.FileSize == 0 {
				t.Fatalf("unexpected image metadata: recipe=%q image=%#v", recipeID, image)
			}
			image.ID = testImageID
			image.RecipeID = recipeID
			image.IsCover = true
			image.UploadedAt = time.Now()
			return image, nil
		},
	}
	handler := NewImageHandler(recipes, store)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/"+testRecipeID+"/images", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", testRecipeID)
	rec := httptest.NewRecorder()

	handler.Upload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(store.data) != 1 {
		t.Fatalf("stored images = %d, want 1", len(store.data))
	}
	for _, data := range store.data {
		if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
			t.Fatalf("stored image is not JPEG: %v", err)
		}
	}
}

func TestServeImageUsesPrivateCaching(t *testing.T) {
	store := &memoryImageStore{data: map[string][]byte{"photo.jpg": []byte("jpeg")}}
	recipes := fakeRecipeRepository{
		getImage: func(context.Context, string) (*repository.RecipeImage, error) {
			return &repository.RecipeImage{ID: testImageID, S3Key: "photo.jpg"}, nil
		},
	}
	handler := NewImageHandler(recipes, store)
	req := httptest.NewRequest(http.MethodGet, "/api/recipe-images/"+testImageID, nil)
	req.SetPathValue("id", testImageID)
	rec := httptest.NewRecorder()

	handler.Serve(rec, req)

	if rec.Code != http.StatusOK || rec.Body.String() != "jpeg" {
		t.Fatalf("status = %d body = %q", rec.Code, rec.Body.String())
	}
	if cache := rec.Header().Get("Cache-Control"); cache != "private, max-age=86400" {
		t.Fatalf("Cache-Control = %q", cache)
	}
}

package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/storage"
)

const (
	maxImageSize   = 10 << 20 // 10 MB
	maxOriginalPx  = 2000
	largePx        = 1200
	mediumPx       = 800
	thumbnailPx    = 300
)

var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

type ImageService struct {
	imageRepo domain.ProductImageRepository
	productRepo domain.ProductRepository
	s3        *storage.S3Client
	log       *zap.Logger
}

func NewImageService(
	imageRepo domain.ProductImageRepository,
	productRepo domain.ProductRepository,
	s3 *storage.S3Client,
	log *zap.Logger,
) *ImageService {
	return &ImageService{
		imageRepo:   imageRepo,
		productRepo: productRepo,
		s3:          s3,
		log:         log,
	}
}

type ImageUploadResult struct {
	Image *domain.ProductImage `json:"image"`
}

// Upload processes and uploads an image for a product.
func (s *ImageService) Upload(ctx context.Context, productID int, fileData io.Reader, contentType string, fileSize int64) (*domain.ProductImage, error) {
	// Validate content type
	if !allowedContentTypes[contentType] {
		return nil, fmt.Errorf("unsupported image format: %s (allowed: JPEG, PNG, WebP)", contentType)
	}

	// Validate file size
	if fileSize > maxImageSize {
		return nil, fmt.Errorf("image too large: %d bytes (max: %d MB)", fileSize, maxImageSize>>20)
	}

	// Check product exists
	_, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Read file data
	data, err := io.ReadAll(fileData)
	if err != nil {
		return nil, fmt.Errorf("read image data: %w", err)
	}

	// Decode image
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// Generate unique prefix
	id := uuid.New().String()
	prefix := fmt.Sprintf("products/%d/%s", productID, id)

	// Determine output format
	ext := ".jpg"
	outputContentType := "image/jpeg"
	if contentType == "image/png" {
		ext = ".png"
		outputContentType = "image/png"
	}

	// Resize and upload all versions
	versions := []struct {
		suffix string
		width  int
	}{
		{"original", maxOriginalPx},
		{"large", largePx},
		{"medium", mediumPx},
		{"thumbnail", thumbnailPx},
	}

	urls := make(map[string]string, 4)
	for _, v := range versions {
		resized := resizeImage(src, v.width)
		encoded, err := encodeImage(resized, contentType)
		if err != nil {
			return nil, fmt.Errorf("encode %s: %w", v.suffix, err)
		}

		key := path.Join(prefix, v.suffix+ext)
		url, err := s.s3.Upload(ctx, key, encoded, outputContentType)
		if err != nil {
			return nil, fmt.Errorf("upload %s: %w", v.suffix, err)
		}
		urls[v.suffix] = url
	}

	// Check if this is the first image — make it main
	count, err := s.imageRepo.CountByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	urlLarge := urls["large"]
	urlMedium := urls["medium"]
	urlThumb := urls["thumbnail"]

	img := &domain.ProductImage{
		ProductID:    productID,
		URL:          urls["original"],
		URLLarge:     &urlLarge,
		URLMedium:    &urlMedium,
		URLThumbnail: &urlThumb,
		S3Key:        prefix,
		IsMain:       count == 0, // first image becomes main
		DisplayOrder: int(count),
	}

	if err := s.imageRepo.Create(ctx, img); err != nil {
		return nil, err
	}

	s.log.Info("image uploaded",
		zap.Int("productID", productID),
		zap.Int("imageID", img.ID),
		zap.String("s3Key", prefix),
	)

	return img, nil
}

// SetMain sets an image as the main image for a product.
func (s *ImageService) SetMain(ctx context.Context, productID int, imageID int) error {
	img, err := s.imageRepo.FindByID(ctx, imageID)
	if err != nil {
		return err
	}
	if img.ProductID != productID {
		return domain.ErrImageNotFound
	}

	return s.imageRepo.SetMain(ctx, productID, imageID)
}

// Delete removes an image from S3 and the database.
func (s *ImageService) Delete(ctx context.Context, imageID int) error {
	img, err := s.imageRepo.FindByID(ctx, imageID)
	if err != nil {
		return err
	}

	// Delete all files in S3 under the prefix
	if err := s.s3.DeletePrefix(ctx, img.S3Key); err != nil {
		s.log.Error("failed to delete s3 files", zap.String("prefix", img.S3Key), zap.Error(err))
		// Continue to delete DB record even if S3 fails
	}

	if err := s.imageRepo.Delete(ctx, imageID); err != nil {
		return err
	}

	s.log.Info("image deleted",
		zap.Int("imageID", imageID),
		zap.String("s3Key", img.S3Key),
	)

	return nil
}

// GetByProductID returns all images for a product.
func (s *ImageService) GetByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error) {
	return s.imageRepo.FindByProductID(ctx, productID)
}

// resizeImage resizes an image to fit within maxWidth, maintaining aspect ratio.
// If the image is already smaller, it returns the original.
func resizeImage(src image.Image, maxWidth int) image.Image {
	bounds := src.Bounds()
	if bounds.Dx() <= maxWidth {
		return src
	}
	return imaging.Resize(src, maxWidth, 0, imaging.Lanczos)
}

// encodeImage encodes an image to bytes in the specified format.
func encodeImage(img image.Image, contentType string) ([]byte, error) {
	var buf bytes.Buffer
	switch {
	case strings.Contains(contentType, "png"):
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	default:
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

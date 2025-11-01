package upload

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image/png"
	"mime/multipart"
	"path/filepath"

	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/storage"
)

type Service struct {
	store storage.Store
	db    *db.Queries
}

func NewService(store storage.Store, db *db.Queries) *Service {
	return &Service{
		store: store,
		db:    db,
	}
}
func (s *Service) SavePDF(username, jobTitle string, file multipart.File) (string, error) {

	// username / jobTitle / pdfs / file
	randomFileName := generateRandomFileName(32)
	pdfPath := filepath.Join(username, jobTitle, "pdfs", randomFileName+".pdf")

	err := s.store.Save(pdfPath, file)
	if err != nil {
		return "", fmt.Errorf("uploadService: failed to save pdf: %w", err)
	}

	return pdfPath, nil
}
func (s *Service) SavePDFImage(username, jobTitle, pdfPath string) (string, error) {

	pdfFile, err := s.store.Read(pdfPath)
	if err != nil {
		return "", fmt.Errorf("uploadService: failed to read pdf: %w", err)
	}

	image, err := common.PDFToImage(pdfFile)
	if err != nil {
		return "", fmt.Errorf("uploadService: failed to convert pdf to image: %w", err)
	}
	var imageBuf bytes.Buffer
	_ = png.Encode(&imageBuf, image)

	// username / jobTitle / images / file
	randomFileName := generateRandomFileName(32)
	imagePath := filepath.Join(username, jobTitle, "images", randomFileName+".png")

	err = s.store.Save(imagePath, &imageBuf)
	if err != nil {
		return "", fmt.Errorf("uploadService: failed to save image: %w", err)
	}

	return imagePath, nil
}

func generateRandomFileName(length int) string {
	randomBytes := make([]byte, length)
	rand.Read(randomBytes)
	return base64.URLEncoding.EncodeToString(randomBytes)
}

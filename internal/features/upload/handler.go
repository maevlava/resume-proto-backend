package upload

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"image/png"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	store storage.Store
}

func NewHandler(store storage.Store) *Handler {
	return &Handler{
		store: store,
	}
}
func (h *Handler) RegisterRoutes(r *http.ServeMux, mws ...middleware.Middleware) {
	r.Handle("/api/v1/uploadpdf", middleware.Chain(http.HandlerFunc(h.UploadPDF), mws...))
}

func (h *Handler) UploadPDF(w http.ResponseWriter, r *http.Request) {
	// user from middleware context
	username, ok := r.Context().Value("username").(string)
	if !ok {
		log.Error().Msg("User not found in context")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	const maxImageSize = 20 * 1024 * 1024
	r.ParseMultipartForm(maxImageSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Error().Msg("Failed to read file from request")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	defer file.Close()

	category := r.FormValue("category")

	// validate pdf
	rawContentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(rawContentType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse media type")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if mediaType != "application/pdf" {
		log.Error().Msg("Invalid media type")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid media type")
		return
	}
	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate random bytes")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	randomFileName := base64.RawURLEncoding.EncodeToString(randomBytes)

	pdfPath := filepath.Join(username, "pdf", category, randomFileName+".pdf")
	log.Info().Msg(pdfPath)
	err = h.store.Save(pdfPath, file)
	if err != nil {
		log.Error().Err(err).Msg("Error saving file")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	imageFile, err := h.store.Read(pdfPath)
	img, err := common.PDFToImage(imageFile)
	if err != nil {
		log.Error().Err(err).Msg("Error converting pdf to image")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		log.Error().Err(err).Msg("Error encoding image to png")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	imagePath := filepath.Join(username, "images", category, randomFileName+".png")
	err = h.store.Save(imagePath, &buf)
	log.Info().Msg(imagePath)
	if err != nil {
		log.Error().Err(err).Msg("Error saving image")
	}

	common.RespondWithJSON(w, http.StatusOK, map[string]string{"pdfPath": pdfPath, "imagePath": imagePath})
}

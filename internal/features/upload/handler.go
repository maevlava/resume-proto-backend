package upload

import (
	"mime"
	"net/http"

	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/server/httperror"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service *service
}

func NewHandler(store storage.Store) *Handler {
	return &Handler{
		service: NewService(store),
	}
}
func (h *Handler) RegisterRoutes(r *http.ServeMux, mws ...middleware.Middleware) {
	use := func(path string, fn httperror.HandlerFunc) {
		route := httperror.Handler(fn)
		route = middleware.Chain(route, mws...)
		r.Handle(path, route)
	}
	use("/api/v1/upload", h.Upload)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) error {
	// user from middleware context
	username, ok := r.Context().Value("username").(string)
	if !ok {
		return httperror.InternalServerError("user not found in context", nil)
	}

	const maxSize = 20 * 1024 * 1024
	if err := r.ParseMultipartForm(maxSize); err != nil {
		return httperror.BadRequest("failed to parse multipart form", err)
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return httperror.BadRequest("missing or invalid file", err)
	}
	defer file.Close()
	jobtitle := r.FormValue("jobtitle")

	// validate pdf
	rawContentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(rawContentType)
	if err != nil {
		return httperror.BadRequest("failed to parse media type", err)
	}
	if mediaType != "application/pdf" {
		return httperror.BadRequest("invalid media type", nil)
	}

	// saves
	pdfPath, err := h.service.SavePDF(username, jobtitle, file)
	if err != nil {
		return httperror.InternalServerError("failed to save pdf", err)
	}
	imagePath, err := h.service.SavePDFImage(username, jobtitle, pdfPath)
	if err != nil {
		return httperror.InternalServerError("failed to save pdf image", err)
	}

	log.Info().Msgf("PDF saved to %s", pdfPath)
	log.Info().Msgf("PDF image saved to %s", imagePath)

	common.RespondWithJSON(w, http.StatusOK, map[string]string{"pdfPath": pdfPath, "imagePath": imagePath})
	return nil
}

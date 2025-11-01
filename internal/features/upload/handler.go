package upload

import (
	"mime"
	"net/http"

	"github.com/maevlava/resume-backend/internal/features/auth"
	"github.com/maevlava/resume-backend/internal/features/resume"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/server/httperror"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service       *Service
	authService   *auth.Service
	resumeService *resume.Service
}

func NewHandler(store storage.Store, db *db.Queries, cfg *config.Config) *Handler {
	return &Handler{
		service:       NewService(store, db),
		authService:   auth.NewService(cfg, db),
		resumeService: resume.NewService(db, store),
	}
}
func (h *Handler) RegisterRoutes(r *http.ServeMux, baseApiPath string, mws ...middleware.Middleware) {
	fullPath := baseApiPath + "/upload"

	use := func(path string, fn httperror.HandlerFunc) {
		route := httperror.Handler(fn)
		route = middleware.Chain(route, mws...)
		r.Handle(fullPath+path, route)
	}
	use("", h.Upload)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return httperror.Unauthorized("upload: failed to get cookie")
	}

	user, err := h.authService.GetSignedInUser(r.Context(), cookie.Value)
	if err != nil {
		return httperror.Unauthorized("upload: failed to get user")
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
	jobTitle := r.FormValue("jobTitle")
	jobDescription := r.FormValue("jobDescription")
	companyName := r.FormValue("companyName")

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
	pdfPath, err := h.service.SavePDF(user.Name, jobTitle, file)
	if err != nil {
		return httperror.InternalServerError("failed to save pdf", err)
	}
	imagePath, err := h.service.SavePDFImage(user.Name, jobTitle, pdfPath)
	if err != nil {
		return httperror.InternalServerError("failed to save pdf image", err)
	}

	resumeID, err := h.resumeService.CreateResume(r.Context(), resume.CreateResumeParams{
		Name:        user.Name,
		UserID:      user.ID,
		Title:       jobTitle,
		Description: jobDescription,
		CompanyName: companyName,
		PdfPath:     pdfPath,
		ImagePath:   imagePath,
	})
	if err != nil {
		return httperror.InternalServerError("failed to create resume", err)
	}
	log.Info().Msgf("resume created: %s", resumeID.String())

	common.RespondWithJSON(w, http.StatusOK, map[string]string{"resumeID": resumeID.String()})
	return nil
}

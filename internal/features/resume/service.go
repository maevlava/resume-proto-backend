package resume

import (
	"context"
	"database/sql"
	"encoding/json/v2"
	"fmt"

	"github.com/google/uuid"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/deepseek"
	"github.com/maevlava/resume-backend/internal/shared/domain"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
)

type Service struct {
	db    *db.Queries
	store storage.Store
}
type CreateResumeParams struct {
	Name        string
	UserID      uuid.UUID
	Title       string
	Description string
	CompanyName string
	ImagePath   string
	PdfPath     string
}

func NewService(db *db.Queries, store storage.Store) *Service {
	return &Service{
		db:    db,
		store: store,
	}
}

func (s *Service) GetResumeByID(ctx context.Context, id string) (*domain.Resume, error) {
	resumeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("GetResumeByID: invalid resume id %w", err)
	}

	dbResume, err := s.db.GetResumeByID(ctx, resumeID)
	if err != nil {
		return nil, fmt.Errorf("GetResumeByID: failed to get resume by id: %w", err)
	}

	resume := domain.Resume{
		ID:             resumeID,
		Username:       dbResume.Name,
		JobTitle:       dbResume.Title,
		JobDescription: dbResume.Description,
		CompanyName:    dbResume.CompanyName,
		PdfPath:        dbResume.PdfPath,
		ImagePath:      dbResume.ImagePath,
	}

	if dbResume.Feedback.Valid {
		var jsonFeedback any
		err = json.Unmarshal([]byte(dbResume.Feedback.String), &jsonFeedback)
		if err != nil {
			return nil, fmt.Errorf("GetResumeByID: failed to unmarshal feedback: %w", err)
		}

		resume.Feedback = jsonFeedback
	}
	log.Info().Msgf("Resume: %v", resume)

	return &resume, nil
}
func (s *Service) GetAllResumesByID(ctx context.Context, resumeID uuid.UUID) ([]*domain.Resume, error) {

	dbResumes, err := s.db.GetResumesByUserID(ctx, resumeID)
	if err != nil {
		return nil, fmt.Errorf("GetResumesByID: failed to get resumes by user: %w", err)
	}

	resumes := make([]*domain.Resume, len(dbResumes))
	for i, dbResume := range dbResumes {
		resumes[i] = &domain.Resume{
			ID:             dbResume.ID,
			Username:       dbResume.Name,
			JobTitle:       dbResume.Title,
			JobDescription: dbResume.Description,
			CompanyName:    dbResume.CompanyName,
			Feedback:       dbResume.Feedback,
			PdfPath:        dbResume.PdfPath,
			ImagePath:      dbResume.ImagePath,
		}
	}

	return resumes, nil
}

func (s *Service) UpdateResumeFeedback(
	ctx context.Context,
	resumeID string,
	feedback deepseek.ChatResponse) (uuid.UUID, error) {

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("updateResumeFeedback: invalid resume id %w", err)
	}

	if len(feedback.Choices) == 0 {
		return uuid.Nil, fmt.Errorf("updateResumeFeedback: no feedback provided")
	}

	feedbackContent := sql.NullString{
		String: feedback.Choices[0].Message.Content,
		Valid:  true,
	}

	updatedResume, err := s.db.UpdateResumeByID(ctx, db.UpdateResumeByIDParams{
		ID:       resumeUUID,
		Feedback: feedbackContent,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("updateResumeFeedback: failed to update resume: %w", err)
	}

	return updatedResume.ID, nil
}
func (s *Service) CreateResume(
	ctx context.Context, params CreateResumeParams) (uuid.UUID, error) {

	newResumeParams := db.CreateResumeParams{
		ID:          uuid.New(),
		UserID:      params.UserID,
		Name:        params.Name,
		Title:       params.Title,
		Description: params.Description,
		CompanyName: params.CompanyName,
		ImagePath:   params.ImagePath,
		PdfPath:     params.PdfPath,
	}

	newResume, err := s.db.CreateResume(ctx, newResumeParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("uploadService: failed to create resume: %w", err)
	}

	return newResume.ID, nil
}

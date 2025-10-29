package ai

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/deepseek"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
)

type Service struct {
	ai    *deepseek.Client
	db    *db.Queries
	store storage.Store
}

func NewService(store storage.Store, ai *deepseek.Client, db *db.Queries) *Service {
	return &Service{
		ai:    ai,
		store: store,
		db:    db,
	}
}

func (s *Service) Analyze(ctx context.Context, id string) (uuid.UUID, error) {

	resumeID, err := uuid.Parse(id)
	log.Info().Msgf("resume id: %s", resumeID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("analyze: invalid resume id %w", err)
	}
	resume, err := s.getResumeByID(ctx, resumeID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("analyze: failed to get resume by id: %w", err)
	}

	pdfFile, err := s.store.Read(resume.PdfPath)
	if err != nil {
		return uuid.Nil, fmt.Errorf("analyze: File not found %w", err)
	}
	defer pdfFile.Close()

	pdfContent, err := common.ExtractPDFText(pdfFile)
	if err != nil {
		return uuid.Nil, fmt.Errorf("analyze: failed to extract pdf text: %w", err)
	}

	content := resumeContent(pdfContent, resume.Title, resume.Description, resume.CompanyName)

	// send to deepseek
	chatRequest := deepseek.ChatRequest{
		Model:       "deepseek-chat",
		Temperature: 0.8,
		Stream:      false,
		Messages: []deepseek.ChatMessage{
			{
				Role:    "system",
				Content: prepareInstructions(),
			},
			{
				Role:    "user",
				Content: content,
			},
		},
	}
	log.Info().Msgf("chat request: %s", chatRequest)

	chatResponse, err := s.ai.Chat(ctx, chatRequest)
	if err != nil {
		return uuid.Nil, fmt.Errorf("analyze: failed to chat with deepseek: %w", err)
	}

	updatedResumeID, err := s.updateResumeFeedback(ctx, resumeID, *chatResponse)
	if err != nil {
		return uuid.Nil, fmt.Errorf("analyze: failed to update resume feedback: %w", err)
	}

	return updatedResumeID, nil
}
func (s *Service) getResumeByID(ctx context.Context, resumeID uuid.UUID) (db.Resume, error) {

	resume, err := s.db.GetResumeByID(ctx, resumeID)
	if err != nil {
		return db.Resume{}, fmt.Errorf("getResumeByID: failed to get resume by id: %w", err)
	}

	return resume, nil
}

func (s *Service) updateResumeFeedback(
	ctx context.Context,
	resumeID uuid.UUID,
	feedback deepseek.ChatResponse) (uuid.UUID, error) {

	if len(feedback.Choices) == 0 {
		return uuid.Nil, fmt.Errorf("updateResumeFeedback: no feedback provided")
	}
	feedbackContent := sql.NullString{
		String: feedback.Choices[0].Message.Content,
		Valid:  true,
	}

	updatedResume, err := s.db.UpdateResumeByID(ctx, db.UpdateResumeByIDParams{
		ID:       resumeID,
		Feedback: feedbackContent,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("updateResumeFeedback: failed to update resume: %w", err)
	}

	return updatedResume.ID, nil
}

const ResponseFormat = `
      interface Feedback {
      overallScore: number; // max 100
      ATS: {
        score: number; // rate based on ATS suitability
        tips: {
          type: "good" | "improve";
          tip: string; // give 3–4 tips
        }[];
      };
      toneAndStyle: {
        score: number; // max 100
        tips: {
          type: "good" | "improve";
          tip: string; // short "title" for the explanation
          explanation: string; // explain in detail here
        }[]; // give 3–4 tips
      };
      content: {
        score: number; // max 100
        tips: {
          type: "good" | "improve";
          tip: string; // short "title" for the explanation
          explanation: string; // explain in detail here
        }[]; // give 3–4 tips
      };
      structure: {
        score: number; // max 100
        tips: {
          type: "good" | "improve";
          tip: string; // short "title" for the explanation
          explanation: string; // explain in detail here
        }[]; // give 3–4 tips
      };
      skills: {
        score: number; // max 100
        tips: {
          type: "good" | "improve";
          tip: string; // short "title" for the explanation
          explanation: string; // explain in detail here
        }[]; // give 3–4 tips
      };
`

func prepareInstructions() string {
	AIBehavior := fmt.Sprintf(`You are an expert in ATS (Applicant Tracking System) and professional resume analysis.

Your task:
- Analyze and rate this resume thoroughly.
- Identify strengths and weaknesses in each section.
- Be honest and critical; low scores are fine if the resume needs improvement.
- Use the job title, description, and company context to refine your evaluation.

Output requirements:
- Respond **only** with a valid JSON object that strictly follows this format: %s
- Do **not** include markdown, code fences (like `+"`"+`json), or any additional text before or after the JSON.
- Do **not** write explanations outside of the JSON object.
- If a field cannot be filled, use a placeholder such as 0 or "N/A".
- The response must begin with '{' and end with '}', forming one complete JSON object.

If your response contains anything other than valid JSON, it will be rejected.`, ResponseFormat)

	return AIBehavior
}
func resumeContent(pdfContent, jobTitle, jobDescription, companyName string) string {
	return fmt.Sprintf(`
	The resume content is as follows:

	%s

	Additional details:
	- Job Title: %s
	- Job Description: %s
	- Company: %s
	`, pdfContent, jobTitle, jobDescription, companyName)
}

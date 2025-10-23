package ai

import (
	"context"
	"fmt"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/deepseek"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
)

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

func Analyze(
	ctx context.Context,
	AI *deepseek.Client,
	store storage.Store,
	request AnalyzeResumeRequest) (*deepseek.ChatResponse, error) {

	// get the pdf file path - extract
	pdfFile, err := store.Read(request.ResumePath)
	log.Info().Msgf("pdf file path: %s", pdfFile)
	if err != nil {
		return nil, fmt.Errorf("file not found")
	}
	defer pdfFile.Close()

	pdfContent, err := common.ExtractPDFText(pdfFile)
	if err != nil {
		return nil, err
	}

	// get job title, description and company name
	content := fmt.Sprintf(`
	The resume content is as follows:

	%s

	Additional details:
	- Job Title: %s
	- Job Description: %s
	- Company: %s
	`, pdfContent, request.JobTitle, request.JobDescription, request.CompanyName)

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

	chatResponse, err := AI.Chat(ctx, chatRequest)
	if err != nil {
		return nil, err
	}

	return chatResponse, nil
}

package domain

import "github.com/google/uuid"

type Resume struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"userId"`
	Username       string    `json:"username"`
	JobTitle       string    `json:"jobTitle"`
	JobDescription string    `json:"jobDescription"`
	CompanyName    string    `json:"companyName"`
	Feedback       any       `json:"feedback"`
	PdfPath        string    `json:"pdfPath"`
	ImagePath      string    `json:"imagePath"`
}

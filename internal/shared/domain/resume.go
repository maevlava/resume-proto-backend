package domain

import "github.com/google/uuid"

type Resume struct {
	ID             uuid.UUID
	Username       string
	JobTitle       string
	JobDescription string
	CompanyName    string
	PdfPath        string
	ImagePath      string
}

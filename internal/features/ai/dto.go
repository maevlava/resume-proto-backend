package ai

type AnalyzeResumeRequest struct {
	ResumePath     string `json:"filePath"`
	CompanyName    string `json:"companyName"`
	JobTitle       string `json:"jobTitle"`
	JobDescription string `json:"jobDescription"`
}

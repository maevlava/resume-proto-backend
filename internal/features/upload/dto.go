package upload

import "mime/multipart"

type UploadRequest struct {
	File           multipart.File
	JobTitle       string `form:"job_title"`
	JobDescription string `form:"job_description"`
	CompanyName    string `form:"company_name"`
}

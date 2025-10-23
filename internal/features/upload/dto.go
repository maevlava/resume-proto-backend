package upload

import "mime/multipart"

type UploadRequest struct {
	File     multipart.File
	Category string `form:"category"`
}

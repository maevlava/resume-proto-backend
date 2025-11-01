-- name: CreateResume :one
INSERT INTO resumes(id, user_id, name, title, description, company_name, feedback, image_path, pdf_path)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
       )
RETURNING *;

-- name: GetResumeByID :one
SELECT * FROM resumes
WHERE resumes.id = $1;

-- name: GetResumesByUserID :many
SELECT * FROM resumes
WHERE resumes.user_id = $1;

-- name: UpdateResumeByID :one
UPDATE resumes
SET
    updated_at = now(),
    feedback = $2
WHERE id = $1
RETURNING *;
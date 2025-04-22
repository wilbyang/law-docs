-- name: CreateDocument :one
INSERT INTO documents (
    title,
    content,
    doc_size,
    created_at,
    updated_at,
    meta,
    status,
    author_id,
    file_path
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
) RETURNING *;

-- name: GetDocumentById :one
SELECT * FROM documents WHERE id = $1;
-- name: GetDocuments :many
SELECT * FROM documents WHERE author_id = $1;

-- name: UpdateDocument :one
UPDATE documents SET title = $2, content = $3, doc_size = $4, updated_at = $5, meta = $6, status = $7, file_path = $8 WHERE id = $1 RETURNING *;
-- name: CreateDocument :one
INSERT INTO documents (
    title,
    content,
    doc_size,
    created_at,
    updated_at,
    meta
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) RETURNING *;
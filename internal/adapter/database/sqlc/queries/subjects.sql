-- name: CreateSubject :one
INSERT INTO subjects (name) VALUES ($1) RETURNING *;

-- name: GetSubject :one
SELECT * FROM subjects WHERE id = $1;

-- name: ListSubjects :many
SELECT * FROM subjects ORDER BY name;

-- name: UpdateSubject :one
UPDATE subjects SET name = $2 WHERE id = $1 RETURNING *;

-- name: DeleteSubject :exec
DELETE FROM subjects WHERE id = $1;

-- name: GetSubjectByName :one
SELECT * FROM subjects WHERE name = $1;

-- name: ListSubjectsByName :many
SELECT *
FROM subjects
WHERE
    name ILIKE '%' || $1 || '%'
ORDER BY name;

-- name: CountSubjects :one
SELECT COUNT(*) FROM subjects;

-- name: CountSubjectsByName :one
SELECT COUNT(*) FROM subjects WHERE name ILIKE '%' || $1 || '%';
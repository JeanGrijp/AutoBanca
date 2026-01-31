-- name: CreateAlternativa :one
INSERT INTO alternativas (question_id, texto_alternativa, correta)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAlternativa :one
SELECT *
FROM alternativas
WHERE id = $1;

-- name: ListAlternativasByQuestion :many
SELECT *
FROM alternativas
WHERE question_id = $1
ORDER BY id;

-- name: UpdateAlternativa :one
UPDATE alternativas
SET
    texto_alternativa = $2,
    correta = $3
WHERE id = $1
RETURNING *;

-- name: DeleteAlternativa :exec
DELETE FROM alternativas
WHERE id = $1;

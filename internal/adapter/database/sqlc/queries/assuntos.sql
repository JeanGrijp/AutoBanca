-- name: CreateAssunto :one
INSERT INTO assuntos (disciplina_id, nome)
VALUES ($1, $2)
RETURNING *;

-- name: GetAssunto :one
SELECT *
FROM assuntos
WHERE id = $1;

-- name: ListAssuntos :many
SELECT *
FROM assuntos
ORDER BY nome;

-- name: ListAssuntosByDisciplina :many
SELECT *
FROM assuntos
WHERE disciplina_id = $1
ORDER BY nome;

-- name: UpdateAssunto :one
UPDATE assuntos
SET
    disciplina_id = $2,
    nome = $3
WHERE id = $1
RETURNING *;

-- name: DeleteAssunto :exec
DELETE FROM assuntos
WHERE id = $1;

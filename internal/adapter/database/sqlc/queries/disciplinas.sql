-- name: CreateDisciplina :one
INSERT INTO disciplinas (nome)
VALUES ($1)
RETURNING *;

-- name: GetDisciplina :one
SELECT *
FROM disciplinas
WHERE id = $1;

-- name: ListDisciplinas :many
SELECT *
FROM disciplinas
ORDER BY nome;

-- name: UpdateDisciplina :one
UPDATE disciplinas
SET nome = $2
WHERE id = $1
RETURNING *;

-- name: DeleteDisciplina :exec
DELETE FROM disciplinas
WHERE id = $1;

-- name: CreateChoice :one
INSERT INTO
    choices (
        question_id,
        choice_text,
        is_correct
    )
VALUES ($1, $2, $3) RETURNING *;

-- name: GetChoice :one
SELECT * FROM choices WHERE id = $1;

-- name: ListChoicesByQuestion :many
SELECT * FROM choices WHERE question_id = $1 ORDER BY id;

-- name: UpdateChoice :one
UPDATE choices
SET
    choice_text = $2,
    is_correct = $3
WHERE
    id = $1 RETURNING *;

-- name: DeleteChoice :exec
DELETE FROM choices WHERE id = $1;
-- name: CreateTopic :one
INSERT INTO topics (subject_id, name) VALUES ($1, $2) RETURNING *;

-- name: GetTopic :one
SELECT * FROM topics WHERE id = $1;

-- name: ListTopics :many
SELECT * FROM topics ORDER BY name;

-- name: ListTopicsBySubject :many
SELECT * FROM topics WHERE subject_id = $1 ORDER BY name;

-- name: UpdateTopic :one
UPDATE topics
SET
    subject_id = $2,
    name = $3
WHERE
    id = $1 RETURNING *;

-- name: DeleteTopic :exec
DELETE FROM topics WHERE id = $1;
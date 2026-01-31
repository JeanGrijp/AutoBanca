-- name: CreateQuestion :one
INSERT INTO
    questions (
        statement,
        year,
        topic_id,
        position,
        level,
        difficulty,
        modality,
        practice_area,
        field_of_study
    )
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
    ) RETURNING *;

-- name: QuestionExistsByStatement :one
SELECT EXISTS (
        SELECT 1
        FROM questions
        WHERE
            statement = $1
    ) AS exists;

-- name: GetQuestion :one
SELECT * FROM questions WHERE id = $1;

-- name: ListQuestions :many
SELECT * FROM questions ORDER BY created_at DESC;

-- name: UpdateQuestion :one
UPDATE questions
SET
    statement = $2,
    year = $3,
    topic_id = $4,
    position = $5,
    level = $6,
    difficulty = $7,
    modality = $8,
    practice_area = $9,
    field_of_study = $10
WHERE
    id = $1 RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;

-- name: ListQuestionsByTopic :many
SELECT *
FROM questions
WHERE
    topic_id = $1
ORDER BY created_at DESC;

-- name: CountQuestions :one
SELECT COUNT(*) FROM questions;

-- name: CountQuestionsByTopic :one
SELECT COUNT(*) FROM questions WHERE topic_id = $1;

-- name: ListQuestionsByYear :many
SELECT * FROM questions WHERE year = $1 ORDER BY created_at DESC;

-- name: CountQuestionsByYear :one
SELECT COUNT(*) FROM questions WHERE year = $1;

-- name: ListQuestionsByLevel :many
SELECT *
FROM questions
WHERE
    level = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByLevel :one
SELECT COUNT(*) FROM questions WHERE level = $1;

-- name: ListQuestionsByModality :many
SELECT *
FROM questions
WHERE
    modality = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByModality :one
SELECT COUNT(*) FROM questions WHERE modality = $1;

-- name: ListQuestionsByPracticeArea :many
SELECT *
FROM questions
WHERE
    practice_area = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByPracticeArea :one
SELECT COUNT(*) FROM questions WHERE practice_area = $1;

-- name: ListQuestionsByFieldOfStudy :many
SELECT *
FROM questions
WHERE
    field_of_study = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByFieldOfStudy :one
SELECT COUNT(*) FROM questions WHERE field_of_study = $1;

-- name: ListQuestionsByYearAndLevel :many
SELECT *
FROM questions
WHERE
    year = $1
    AND level = $2
ORDER BY created_at DESC;

-- name: CountQuestionsByYearAndLevel :one
SELECT COUNT(*) FROM questions WHERE year = $1 AND level = $2;

-- name: ListQuestionsByFilters :many
SELECT *
FROM questions
WHERE
    (sqlc.narg('year')::INT IS NULL OR year = sqlc.narg('year'))
    AND (sqlc.narg('level')::TEXT IS NULL OR level = sqlc.narg('level'))
    AND (sqlc.narg('difficulty')::TEXT IS NULL OR difficulty = sqlc.narg('difficulty'))
    AND (sqlc.narg('modality')::TEXT IS NULL OR modality = sqlc.narg('modality'))
    AND (sqlc.narg('practice_area')::TEXT IS NULL OR practice_area = sqlc.narg('practice_area'))
    AND (sqlc.narg('field_of_study')::TEXT IS NULL OR field_of_study = sqlc.narg('field_of_study'))
    AND (sqlc.narg('topic_id')::UUID IS NULL OR topic_id = sqlc.narg('topic_id'))
    AND (sqlc.narg('position')::TEXT IS NULL OR position = sqlc.narg('position'))
    AND (sqlc.narg('questions_count')::INT IS NULL OR TRUE)
ORDER BY created_at DESC;

-- name: CountQuestionsByFilters :one
SELECT COUNT(*)
FROM questions
WHERE
    (sqlc.narg('year')::INT IS NULL OR year = sqlc.narg('year'))
    AND (sqlc.narg('level')::TEXT IS NULL OR level = sqlc.narg('level'))
    AND (sqlc.narg('difficulty')::TEXT IS NULL OR difficulty = sqlc.narg('difficulty'))
    AND (sqlc.narg('modality')::TEXT IS NULL OR modality = sqlc.narg('modality'))
    AND (sqlc.narg('practice_area')::TEXT IS NULL OR practice_area = sqlc.narg('practice_area'))
    AND (sqlc.narg('field_of_study')::TEXT IS NULL OR field_of_study = sqlc.narg('field_of_study'))
    AND (sqlc.narg('topic_id')::UUID IS NULL OR topic_id = sqlc.narg('topic_id'))
    AND (sqlc.narg('position')::TEXT IS NULL OR position = sqlc.narg('position'));

-- name: ListQuestionsByFiltersWithChoices :many
SELECT 
    q.id, q.statement, q.year, q.position, q.level,
    q.difficulty, q.modality, q.practice_area, q.field_of_study,
    c.id as choice_id, c.choice_text, c.is_correct
FROM questions q
LEFT JOIN choices c ON q.id = c.question_id
WHERE
    (sqlc.narg('year')::INT IS NULL OR q.year = sqlc.narg('year'))
    AND (sqlc.narg('level')::TEXT IS NULL OR q.level = sqlc.narg('level'))
    AND (sqlc.narg('difficulty')::TEXT IS NULL OR q.difficulty = sqlc.narg('difficulty'))
    AND (sqlc.narg('modality')::TEXT IS NULL OR q.modality = sqlc.narg('modality'))
    AND (sqlc.narg('practice_area')::TEXT IS NULL OR q.practice_area = sqlc.narg('practice_area'))
    AND (sqlc.narg('field_of_study')::TEXT IS NULL OR q.field_of_study = sqlc.narg('field_of_study'))
    AND (sqlc.narg('topic_id')::UUID IS NULL OR q.topic_id = sqlc.narg('topic_id'))
    AND (sqlc.narg('position')::TEXT IS NULL OR q.position = sqlc.narg('position'))
ORDER BY q.created_at DESC;

-- name: GetQuestionsForExam :many
SELECT 
    q.id, q.statement, q.year, q.position, q.level,
    q.difficulty, q.modality, q.practice_area, q.field_of_study,
    t.name as topic_name, s.name as subject_name
FROM questions q
JOIN topics t ON q.topic_id = t.id
JOIN subjects s ON t.subject_id = s.id
WHERE 
    s.id = $1 
    AND (sqlc.narg('topic_id')::uuid IS NULL OR q.topic_id = sqlc.narg('topic_id'))
    AND (sqlc.narg('position')::text IS NULL OR q.position = sqlc.narg('position'))
    AND (sqlc.narg('level')::text IS NULL OR q.level = sqlc.narg('level'))
    AND (sqlc.narg('difficulty')::text IS NULL OR q.difficulty = sqlc.narg('difficulty'))
ORDER BY RANDOM()
LIMIT $2;

-- name: CountQuestionsForExam :one
SELECT COUNT(*)
FROM questions q
JOIN topics t ON q.topic_id = t.id
JOIN subjects s ON t.subject_id = s.id
WHERE 
    s.id = $1 
    AND (sqlc.narg('topic_id')::uuid IS NULL OR q.topic_id = sqlc.narg('topic_id'))
    AND (sqlc.narg('position')::text IS NULL OR q.position = sqlc.narg('position'))
    AND (sqlc.narg('level')::text IS NULL OR q.level = sqlc.narg('level'))
    AND (sqlc.narg('difficulty')::text IS NULL OR q.difficulty = sqlc.narg('difficulty'));
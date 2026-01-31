-- name: CreateQuestion :one
INSERT INTO
    questions (
        enunciado,
        ano,
        assunto_id,
        instituicao,
        cargo,
        nivel,
        dificuldade,
        modalidade,
        area_atuacao,
        area_formacao
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
        $9,
        $10
    ) RETURNING *;

-- name: GetQuestion :one
SELECT * FROM questions WHERE id = $1;

-- name: ListQuestions :many
SELECT * FROM questions ORDER BY created_at DESC;

-- name: UpdateQuestion :one
UPDATE questions
SET
    enunciado = $2,
    ano = $3,
    assunto_id = $4,
    instituicao = $5,
    cargo = $6,
    nivel = $7,
    dificuldade = $8,
    modalidade = $9,
    area_atuacao = $10,
    area_formacao = $11
WHERE
    id = $1 RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;

-- name: ListQuestionsByAssunto :many
SELECT *
FROM questions
WHERE
    assunto_id = $1
ORDER BY created_at DESC;

-- name: CountQuestions :one
SELECT COUNT(*) FROM questions;

-- name: CountQuestionsByAssunto :one
SELECT COUNT(*) FROM questions WHERE assunto_id = $1;

-- name: ListQuestionsByInstituicao :many
SELECT *
FROM questions
WHERE
    instituicao = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByInstituicao :one
SELECT COUNT(*) FROM questions WHERE instituicao = $1;

-- name: ListQuestionsByDificuldadeAndInstituicao :many
SELECT *
FROM questions
WHERE
    dificuldade = $1
    AND instituicao = $2
ORDER BY created_at DESC;

-- name: CountQuestionsByDificuldadeAndInstituicao :one
SELECT COUNT(*)
FROM questions
WHERE
    dificuldade = $1
    AND instituicao = $2;

-- name: ListQuestionsByAno :many
SELECT * FROM questions WHERE ano = $1 ORDER BY created_at DESC;

-- name: CountQuestionsByAno :one
SELECT COUNT(*) FROM questions WHERE ano = $1;

-- name: ListQuestionsByNivel :many
SELECT *
FROM questions
WHERE
    nivel = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByNivel :one
SELECT COUNT(*) FROM questions WHERE nivel = $1;

-- name: ListQuestionsByModalidade :many
SELECT *
FROM questions
WHERE
    modalidade = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByModalidade :one
SELECT COUNT(*) FROM questions WHERE modalidade = $1;

-- name: ListQuestionsByAreaAtuacao :many
SELECT *
FROM questions
WHERE
    area_atuacao = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByAreaAtuacao :one
SELECT COUNT(*) FROM questions WHERE area_atuacao = $1;

-- name: ListQuestionsByAreaFormacao :many
SELECT *
FROM questions
WHERE
    area_formacao = $1
ORDER BY created_at DESC;

-- name: CountQuestionsByAreaFormacao :one
SELECT COUNT(*) FROM questions WHERE area_formacao = $1;

-- name: ListQuestionsByInstituicaoAndCargo :many
SELECT *
FROM questions
WHERE
    instituicao = $1
    AND cargo = $2
ORDER BY created_at DESC;

-- name: CountQuestionsByInstituicaoAndCargo :one
SELECT COUNT(*) FROM questions WHERE instituicao = $1 AND cargo = $2;

-- name: ListQuestionsByAnoAndNivel :many
SELECT *
FROM questions
WHERE
    ano = $1
    AND nivel = $2
ORDER BY created_at DESC;

-- name: CountQuestionsByAnoAndNivel :one
SELECT COUNT(*) FROM questions WHERE ano = $1 AND nivel = $2;

-- name: ListQuestionsByFilters :many
SELECT *
FROM questions
WHERE
    (sqlc.narg('instituicao')::TEXT IS NULL OR instituicao = sqlc.narg('instituicao'))
    AND (sqlc.narg('ano')::INT IS NULL OR ano = sqlc.narg('ano'))
    AND (sqlc.narg('nivel')::TEXT IS NULL OR nivel = sqlc.narg('nivel'))
    AND (sqlc.narg('dificuldade')::TEXT IS NULL OR dificuldade = sqlc.narg('dificuldade'))
    AND (sqlc.narg('modalidade')::TEXT IS NULL OR modalidade = sqlc.narg('modalidade'))
    AND (sqlc.narg('area_atuacao')::TEXT IS NULL OR area_atuacao = sqlc.narg('area_atuacao'))
    AND (sqlc.narg('area_formacao')::TEXT IS NULL OR area_formacao = sqlc.narg('area_formacao'))
    AND (sqlc.narg('assunto_id')::UUID IS NULL OR assunto_id = sqlc.narg('assunto_id'))
    AND (sqlc.narg('cargo')::TEXT IS NULL OR cargo = sqlc.narg('cargo'))
ORDER BY created_at DESC;

-- name: CountQuestionsByFilters :one
SELECT COUNT(*)
FROM questions
WHERE
    (sqlc.narg('instituicao')::TEXT IS NULL OR instituicao = sqlc.narg('instituicao'))
    AND (sqlc.narg('ano')::INT IS NULL OR ano = sqlc.narg('ano'))
    AND (sqlc.narg('nivel')::TEXT IS NULL OR nivel = sqlc.narg('nivel'))
    AND (sqlc.narg('dificuldade')::TEXT IS NULL OR dificuldade = sqlc.narg('dificuldade'))
    AND (sqlc.narg('modalidade')::TEXT IS NULL OR modalidade = sqlc.narg('modalidade'))
    AND (sqlc.narg('area_atuacao')::TEXT IS NULL OR area_atuacao = sqlc.narg('area_atuacao'))
    AND (sqlc.narg('area_formacao')::TEXT IS NULL OR area_formacao = sqlc.narg('area_formacao'))
    AND (sqlc.narg('assunto_id')::UUID IS NULL OR assunto_id = sqlc.narg('assunto_id'))
    AND (sqlc.narg('cargo')::TEXT IS NULL OR cargo = sqlc.narg('cargo'));
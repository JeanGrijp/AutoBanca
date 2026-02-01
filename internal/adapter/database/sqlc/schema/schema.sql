-- Ativa a extensão para geração de UUID (caso não esteja ativa)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Subjects table
CREATE TABLE subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Topics table (related to subjects)
CREATE TABLE topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    subject_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    CONSTRAINT fk_subject FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE CASCADE,
    UNIQUE (subject_id, name),
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Questions table
CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    statement TEXT NOT NULL,
    year INT NOT NULL,
    topic_id UUID NOT NULL,
    position VARCHAR(100),
    level VARCHAR(20), -- Suggested: Superior, Médio, Fundamental
    difficulty VARCHAR(20), -- Suggested: Fácil, Média, Difícil
    modality VARCHAR(20), -- Suggested: Múltipla Escolha, Certo/Errado
    practice_area VARCHAR(50),
    field_of_study VARCHAR(50),
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_topic FOREIGN KEY (topic_id) REFERENCES topics (id)
);

-- 4. Choices table
CREATE TABLE choices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    question_id UUID NOT NULL,
    choice_text TEXT NOT NULL,
    is_correct BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_question FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE
);

-- Índices para acelerar a geração automática de provas (filtros comuns)
CREATE INDEX idx_questions_topic_id ON questions (topic_id);

CREATE INDEX idx_questions_level_difficulty ON questions (level, difficulty);

CREATE INDEX idx_questions_practice_area ON questions (practice_area);

CREATE INDEX idx_questions_field_of_study ON questions (field_of_study);

CREATE INDEX idx_choices_question_id ON choices (question_id);
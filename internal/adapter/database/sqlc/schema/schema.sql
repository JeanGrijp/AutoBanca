-- Ativa a extensão para geração de UUID (caso não esteja ativa)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Tabela de Disciplinas
CREATE TABLE disciplinas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    nome VARCHAR(100) NOT NULL UNIQUE
);

-- 2. Tabela de Assuntos (Relacionada a Disciplinas)
CREATE TABLE assuntos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    disciplina_id UUID NOT NULL,
    nome VARCHAR(100) NOT NULL,
    CONSTRAINT fk_disciplina FOREIGN KEY (disciplina_id) REFERENCES disciplinas (id) ON DELETE CASCADE
);

-- 3. Tabela de Questões
CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    enunciado TEXT NOT NULL,
    ano INT NOT NULL,
    assunto_id UUID NOT NULL,
    instituicao VARCHAR(100),
    cargo VARCHAR(100),
    nivel VARCHAR(20), -- Sugestão: Superior, Médio, Fundamental
    dificuldade VARCHAR(20), -- Sugestão: Fácil, Média, Difícil
    modalidade VARCHAR(20), -- Sugestão: Múltipla Escolha, Certo/Errado
    area_atuacao VARCHAR(50),
    area_formacao VARCHAR(50),
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_assunto FOREIGN KEY (assunto_id) REFERENCES assuntos (id)
);

-- 4. Tabela de Alternativas
CREATE TABLE alternativas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    question_id UUID NOT NULL,
    texto_alternativa TEXT NOT NULL,
    correta BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_question FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE
);

-- Índices para acelerar a geração automática de provas (filtros comuns)
CREATE INDEX idx_questions_assunto ON questions (assunto_id);

CREATE INDEX idx_questions_nivel_dificuldade ON questions (nivel, dificuldade);

CREATE INDEX idx_questions_area_atuacao ON questions (area_atuacao);

CREATE INDEX idx_questions_area_formacao ON questions (area_formacao);

CREATE INDEX idx_alternativas_question_id ON alternativas (question_id);
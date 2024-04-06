-- Enable the pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS reports (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES oauth2_authorized(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    report_type TEXT NOT NULL,
    report TEXT NOT NULL,
    embedding VECTOR
);
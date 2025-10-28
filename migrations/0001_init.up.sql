-- Таблица изображений
CREATE TABLE IF NOT EXISTS images (
        id UUID PRIMARY KEY,
        original_path TEXT NOT NULL,
        uploaded_at TIMESTAMP DEFAULT NOW()
);

-- Таблица задач на обработку изображений
CREATE TYPE task_status AS ENUM ('queued', 'processing', 'completed', 'failed');
CREATE TYPE task_type AS ENUM ('resize', 'thumbnail', 'watermark');

CREATE TABLE IF NOT EXISTS tasks (
        id SERIAL PRIMARY KEY,
        image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
        processed_path TEXT,
        type task_type NOT NULL,
        status task_status NOT NULL DEFAULT 'queued',
        created_at TIMESTAMP DEFAULT NOW()
);

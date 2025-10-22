CREATE TABLE IF NOT EXISTS canvas (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS layers (
    id VARCHAR(255) NOT NULL,
    canvas_id VARCHAR(255) NOT NULL REFERENCES canvas(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    visible BOOLEAN NOT NULL DEFAULT TRUE,
    locked BOOLEAN NOT NULL DEFAULT FALSE,
    layer_order INT NOT NULL,
    PRIMARY KEY (id, canvas_id)
);

CREATE TABLE IF NOT EXISTS shapes (
    id SERIAL PRIMARY KEY,
    canvas_id VARCHAR(255) NOT NULL REFERENCES canvas(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    layer_number INT NOT NULL,
    type VARCHAR(50) NOT NULL,
    color VARCHAR(50),
    stroke_width FLOAT,
    attributes JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_layers_canvas_id ON layers(canvas_id);
CREATE INDEX IF NOT EXISTS idx_shapes_canvas_id ON shapes(canvas_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_canvas_updated_at
BEFORE UPDATE ON canvas
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

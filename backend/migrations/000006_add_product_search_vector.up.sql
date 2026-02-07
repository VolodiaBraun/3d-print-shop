-- Add tsvector column for full-text search with Russian configuration
ALTER TABLE products ADD COLUMN search_vector tsvector;

-- Create GIN index for fast FTS
CREATE INDEX idx_products_search_vector ON products USING GIN(search_vector);

-- Populate search_vector for existing rows
UPDATE products SET search_vector =
    setweight(to_tsvector('russian', coalesce(name, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(description, '')), 'B');

-- Trigger function to auto-update search_vector on insert/update
CREATE OR REPLACE FUNCTION products_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('russian', coalesce(NEW.name, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_products_search_vector
    BEFORE INSERT OR UPDATE OF name, description ON products
    FOR EACH ROW
    EXECUTE FUNCTION products_search_vector_update();

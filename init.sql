CREATE EXTENSION "uuid-ossp";

CREATE TYPE chair_type AS ENUM ('abc', 'cde');

CREATE TABLE main (
                                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                    title TEXT NOT NULL,
                                    sub_id INTEGER,
                                    sub_obj JSONB,
                                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                    deleted_at TIMESTAMPTZ
);

CREATE TABLE tools (
                                     id SERIAL PRIMARY KEY,
                                     title TEXT NOT NULL,
                                     description TEXT,
                                     main_id UUID NOT NULL REFERENCES main(id) ON DELETE CASCADE,
                                     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                     deleted_at TIMESTAMPTZ
);

CREATE TABLE tables (
                                      id SERIAL PRIMARY KEY,
                                      name TEXT NOT NULL,
                                      main_id UUID NOT NULL REFERENCES main(id) ON DELETE CASCADE,
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      deleted_at TIMESTAMPTZ
);

CREATE TABLE chairs (
                                      id SERIAL PRIMARY KEY,
                                      name TEXT NOT NULL,
                                      type chair_type NOT NULL,
                                      main_id UUID NOT NULL REFERENCES main(id) ON DELETE CASCADE,
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      deleted_at TIMESTAMPTZ
);


CREATE OR REPLACE FUNCTION propagate_main_sub_obj() RETURNS trigger AS $$
BEGIN
    DELETE FROM tools WHERE main_id = NEW.id;
    DELETE FROM tables WHERE main_id = NEW.id;
    DELETE FROM chairs WHERE main_id = NEW.id;

    IF NEW.sub_obj ? 'tools' AND jsonb_typeof(NEW.sub_obj->'tools') = 'array' THEN
        INSERT INTO tools(title, description, main_id, created_at, updated_at, deleted_at)
        SELECT
            elem->>'title',
            elem->>'description',
            NEW.id,
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at
        FROM jsonb_array_elements(NEW.sub_obj->'tools') AS elem;
    END IF;

    IF NEW.sub_obj ? 'tables' AND jsonb_typeof(NEW.sub_obj->'tables') = 'array' THEN
        INSERT INTO tables(name, main_id, created_at, updated_at, deleted_at)
        SELECT
            elem->>'name',
            NEW.id,
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at
        FROM jsonb_array_elements(NEW.sub_obj->'tables') AS elem;
    END IF;

    IF NEW.sub_obj ? 'chairs' AND jsonb_typeof(NEW.sub_obj->'chairs') = 'array' THEN
        INSERT INTO chairs(name, type, main_id, created_at, updated_at, deleted_at)
        SELECT
            elem->>'name',
            (elem->>'type')::chair_type,
            NEW.id,
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at
        FROM jsonb_array_elements(NEW.sub_obj->'chairs') AS elem;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS propagate_main_sub_obj_insert ON main;
CREATE TRIGGER propagate_main_sub_obj_insert
    AFTER INSERT ON main
    FOR EACH ROW EXECUTE FUNCTION propagate_main_sub_obj();

DROP TRIGGER IF EXISTS propagate_main_sub_obj_update ON main;
CREATE TRIGGER propagate_main_sub_obj_update
    AFTER UPDATE OF sub_obj ON main
    FOR EACH ROW EXECUTE FUNCTION propagate_main_sub_obj();

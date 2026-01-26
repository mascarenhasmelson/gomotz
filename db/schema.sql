
-- CREATE TABLE groups (
--     id SERIAL PRIMARY KEY,
--     name TEXT NOT NULL UNIQUE,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );
DROP TABLE IF EXISTS services;
CREATE TABLE services (
  id SERIAL PRIMARY KEY,
  -- group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  service_name VARCHAR(100) NOT NULL,
  local_ip INET NOT NULL,
  local_port INTEGER NOT NULL CHECK (local_port BETWEEN 1 AND 65535),
  remote_ip INET NOT NULL,
  remote_port INTEGER NOT NULL CHECK (remote_port BETWEEN 1 AND 65535),
  online BOOLEAN DEFAULT FALSE,  
  last_seen TIMESTAMPTZ DEFAULT NOW(),
  pid INTEGER,
  created_at TIMESTAMPTZ DEFAULT NOW()
  -- UNIQUE(service_name, group_id)
);

CREATE OR REPLACE FUNCTION notify_service_change()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('service_change', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- trigger for notify_service_change after insert update, delete
CREATE TRIGGER service_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON services
FOR EACH ROW
EXECUTE FUNCTION notify_service_change();

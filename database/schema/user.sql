CREATE TABLE IF NOT EXISTS oauth2_authorized(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_type TEXT NOT NULL,
    scope TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_in INTEGER NOT NULL,
    access_token TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(access_token),
    UNIQUE(user_id)
);

-- Create the trigger that sends a notification every time a new
-- user is added into the authorizedUser table
CREATE OR REPLACE FUNCTION notify_new_user()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM pg_notify('new_users', NEW.id::text);
    RETURN NULL;
END $$;

CREATE OR REPLACE TRIGGER after_insert_user
	AFTER INSERT ON oauth2_authorized
	FOR EACH ROW EXECUTE FUNCTION notify_new_user();

CREATE TABLE IF NOT EXISTS oauth2_authorizing(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    csrftoken TEXT NOT NULL,
    code TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(csrftoken)
);
CREATE TABLE IF NOT EXISTS oauth2_authorized(
    id bigserial primary key not null,
    user_id TEXT not null,
    token_type TEXT not null,
    scope TEXT not null,
    refresh_token TEXT not null,
    expires_in bigint not null default 0,
    access_token TEXT not null,
    created_at TIMESTAMP not null DEFAULT NOW(),
    UNIQUE(access_token),
    UNIQUE(user_id)
);

-- Create the trigger that sends a notification every time a new
-- user is added into the authorizedUser table.
-- It sends the access_token as payload.
-- COMMENTED: it's demanded to the application layer.
/*
CREATE OR REPLACE FUNCTION notify_new_user()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM pg_notify('new_users', NEW.access_token);
    RETURN NULL;
END $$;


CREATE OR REPLACE TRIGGER after_insert_user
	AFTER INSERT ON oauth2_authorized
	FOR EACH ROW EXECUTE FUNCTION notify_new_user();
*/

CREATE TABLE IF NOT EXISTS oauth2_authorizing(
    id bigserial primary key not null,
    csrftoken TEXT not null,
    code TEXT not null,
    created_at TIMESTAMP not null DEFAULT NOW(),
    UNIQUE(csrftoken)
);
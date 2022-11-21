DO $$
DECLARE
    column_exist boolean := false;
BEGIN
SELECT count(*) != 0 INTO column_exist
    FROM information_schema.columns
    WHERE table_name = 'oauthapps'
    AND column_name = 'mattermostappid';
IF column_exist THEN
    ALTER TABLE OAuthApps ALTER COLUMN MattermostAppID DROP NOT NULL;
    ALTER TABLE OAuthApps ALTER COLUMN MattermostAppID DROP DEFAULT;
END IF;
END $$;

SET @preparedStatement = (SELECT IF(
    (
        SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
        WHERE table_name = 'OAuthApps'
        AND table_schema = DATABASE()
        AND column_name = 'MattermostAppID'
    ) > 0,
    'ALTER TABLE OAuthApps DROP COLUMN MattermostAppID;',
    'SELECT 1'
));

PREPARE alterIfExists FROM @preparedStatement;
EXECUTE alterIfExists;
DEALLOCATE PREPARE alterIfExists;

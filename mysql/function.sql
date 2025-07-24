-- Аналог UUID_TO_BIN()
CREATE FUNCTION UUID_TO_BIN(uuid CHAR(36))
RETURNS BINARY(16)
DETERMINISTIC
BEGIN
    RETURN UNHEX(REPLACE(uuid, '-', ''));
END

-- Аналог BIN_TO_UUID()
CREATE FUNCTION UUID_TO_STRING(bin BINARY(16))
RETURNS CHAR(36)
DETERMINISTIC
BEGIN
    DECLARE hex CHAR(32);
    SET hex = HEX(bin);
    RETURN LOWER(CONCAT(
        SUBSTR(hex, 1, 8), '-',
        SUBSTR(hex, 9, 4), '-',
        SUBSTR(hex, 13, 4), '-',
        SUBSTR(hex, 17, 4), '-',
        SUBSTR(hex, 21, 12)
    ));
END
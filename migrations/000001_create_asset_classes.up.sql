CREATE TABLE asset_classes (
    id   SMALLSERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE
);

INSERT INTO asset_classes (name) VALUES
    ('equity'),
    ('etf'),
    ('crypto'),
    ('commodity'),
    ('forex'),
    ('index');

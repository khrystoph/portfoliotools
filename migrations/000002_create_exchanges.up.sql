CREATE TABLE exchanges (
    id        SERIAL PRIMARY KEY,
    name      VARCHAR(100) NOT NULL,
    acronym   VARCHAR(20),
    mic_code  VARCHAR(10) UNIQUE,
    timezone  VARCHAR(50) NOT NULL DEFAULT 'America/New_York',
    country   VARCHAR(2)  NOT NULL DEFAULT 'US'
);

CREATE DATABASE werd;
\c werd;

-- User management
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE sessions (
    session_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    session_token VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

-- Word management
CREATE TABLE words (
    word_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    word VARCHAR(255) NOT NULL,
    rarity INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE pullTracking (
    pullTime_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    nextPullTime TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
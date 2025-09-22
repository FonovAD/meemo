CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,
    first_name    VARCHAR(255)        NOT NULL,
    last_name     VARCHAR(255)        NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_salt VARCHAR(255)        NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE files
(
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT       NOT NULL,
    original_name VARCHAR(500) NOT NULL,
    mime_type     VARCHAR(255) NOT NULL,
    size_in_bytes BIGINT       NOT NULL,
    s3_bucket     VARCHAR(255) NOT NULL,
    s3_key        VARCHAR(500) NOT NULL,
    status        INTEGER      NOT NULL    DEFAULT 0,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_public     BOOLEAN      NOT NULL    DEFAULT false,

    CONSTRAINT fk_files_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_files_user_id ON files (user_id);
CREATE INDEX idx_files_status ON files (status);
CREATE INDEX idx_files_created_at ON files (created_at);
CREATE INDEX idx_files_is_public ON files (is_public);
CREATE INDEX idx_users_email ON users (email);

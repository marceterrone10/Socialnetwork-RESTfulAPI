CREATE TABLE IF NOT EXISTS roles (
    id bigserial PRIMARY KEY,
    name varchar(255) NOT NULL,
    description text NOT NULL,
    level int NOT NULL DEFAULT 0,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description, level) 
VALUES (
    'user',
    'A user can create posts and comments',
    1
);

INSERT INTO roles (name, description, level) 
VALUES (
    'moderator',
    'A moderator can update other users posts',
    2
);

INSERT INTO roles (name, description, level) 
VALUES (
    'admin',
    'A admin can update and delete any posts and comments',
    3
);

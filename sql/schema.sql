DROP TABLE IF EXISTS user;

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    vk_id bigint NOT NULL CONSTRAINT user_pk PRIMARY KEY,
    carma int NOT NULL default 0,
    name citext,
    surname citext,
    photo_url text
);
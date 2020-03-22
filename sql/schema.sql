DROP TABLE IF EXISTS user;

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE users (
    vk_id bigint NOT NULL CONSTRAINT user_pk PRIMARY KEY,
    carma int NOT NULL default 0,
    name citext,
    surname citext,
    photo_url text
);

CREATE TYPE feedback AS ENUM ('ls', 'comments', 'other');
CREATE TYPE ad_status AS ENUM ('offer', 'chosen', 'closed');

CREATE TABLE ad (
    ad_id bigserial CONSTRAINT ad_pk PRIMARY KEY,
    author_id bigint,
    CONSTRAINT ad_author FOREIGN KEY (author_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    header citext,
    text citext,
    region citext,
    district citext,
    is_auction boolean,
    feedback_type feedback,
    extra_field citext,
    creation_datetime TIMESTAMP WITH TIME ZONE default now(),
    geo_position geography,
    status ad_status DEFAULT 'offer',
    category citext,    -- mb change for enum of categories too
    comments_count int DEFAULT 0
);

CREATE TABLE ad_photos (
    ad_photos_id bigserial CONSTRAINT ad_photos_pk PRIMARY KEY ,
    ad_id bigint,
    CONSTRAINT ad_photos_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    photo_url text
);
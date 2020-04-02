DROP TABLE IF EXISTS deal;
DROP TABLE IF EXISTS ad_subscribers;
DROP TABLE IF EXISTS ad_photos;
DROP TABLE IF EXISTS ad;
DROP TABLE IF EXISTS users;


CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE users (
    vk_id bigint NOT NULL CONSTRAINT user_pk PRIMARY KEY,
    carma int NOT NULL default 0,
    name citext,
    surname citext,
    photo_url text
);

DROP TYPE IF EXISTS feedback;
DROP TYPE IF EXISTS ad_status;
CREATE TYPE feedback AS ENUM ('ls', 'comments', 'other');
CREATE TYPE ad_status AS ENUM ('offer', 'chosen', 'closed', 'aborted');

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
    lat float,
    long float,
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

CREATE TABLE ad_subscribers (
    ad_subscribers_id bigserial CONSTRAINT ad_subscribers_pk PRIMARY KEY ,
    ad_id bigint,
    CONSTRAINT ad_subscribers_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    subscriber_id bigint,
    CONSTRAINT ad_subscribers_user FOREIGN KEY (subscriber_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    CONSTRAINT ad_subscriber_unique UNIQUE (ad_id, subscriber_id)
);

DROP TYPE IF EXISTS deal_status;
CREATE TYPE deal_status AS ENUM ('open', 'success');

CREATE TABLE deal (
    deal_id bigserial CONSTRAINT deal_pk PRIMARY KEY,
    ad_id bigint,
    CONSTRAINT deal_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    subscriber_id bigint,
    status deal_status DEFAULT 'open',
    CONSTRAINT deal_ad_subscriber_unique UNIQUE (ad_id, subscriber_id),
    CONSTRAINT deal_user FOREIGN KEY (subscriber_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION
);

CREATE OR REPLACE FUNCTION make_deal(ad_id_to_insert INT, subscriber_id_to_insert INT) RETURNS INT AS $$
    DECLARE d_id INT;
    BEGIN
        INSERT INTO deal (ad_id, subscriber_id) VALUES (ad_id_to_insert, subscriber_id_to_insert)
         RETURNING deal_id INTO d_id;
        UPDATE ad SET status = 'chosen' WHERE ad_id = ad_id_to_insert;
        return d_id;
    END;
    $$ LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION close_deal_success(deal_id_to_upd INT) RETURNS void AS $$
    DECLARE _ad_id INT;
    BEGIN
        UPDATE deal SET status = 'success' WHERE deal_id = deal_id_to_upd RETURNING ad_id INTO _ad_id;
        UPDATE ad SET status = 'closed' WHERE ad_id = _ad_id;
        DELETE FROM ad_subscribers WHERE ad_id = _ad_id;
    END;
    $$ LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION close_deal_fail_by_author(deal_id_to_cls INT) RETURNS void AS $$
    DECLARE _ad_id INT;
    BEGIN
        _ad_id := (SELECT ad_id FROM deal WHERE deal_id = deal_id_to_cls);
        UPDATE ad SET status = 'offer' WHERE ad_id = _ad_id;
        DELETE FROM deal WHERE deal_id = deal_id_to_cls;
    END;
$$ LANGUAGE 'plpgsql';


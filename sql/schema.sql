DROP TABLE IF EXISTS deal;
DROP TABLE IF EXISTS ad_subscribers;
DROP TABLE IF EXISTS ad_photos;
DROP TABLE IF EXISTS ad;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS ad_view;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS comment;

DROP FUNCTION IF EXISTS make_deal;
DROP FUNCTION IF EXISTS close_deal_success;
DROP FUNCTION IF EXISTS close_deal_fail_by_author;
DROP TRIGGER IF EXISTS update_comments_count ON comment;
DROP FUNCTION IF EXISTS update_comments_count;
drop trigger if exists ad_view_create on ad;
drop function if exists ad_view_create();


DROP INDEX IF EXISTS ad_geos;


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
    comments_count int DEFAULT 0,
    hidden BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE ad_view (
    ad_view_id bigserial CONSTRAINT ad_view_pk PRIMARY KEY ,
    ad_id bigint,
    CONSTRAINT ad_view_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    views_count bigint,
    CONSTRAINT ad_view_unique UNIQUE (ad_id)
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
--         DELETE FROM ad_subscribers WHERE ad_id = _ad_id; TODO: dunno if needed
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

CREATE OR REPLACE FUNCTION close_deal_fail_by_subscriber(deal_id_to_cls INT) RETURNS void AS $$
DECLARE _ad_id INT;
BEGIN
    _ad_id := (SELECT ad_id FROM deal WHERE deal_id = deal_id_to_cls);
    UPDATE ad SET status = 'aborted' WHERE ad_id = _ad_id;
    DELETE FROM deal WHERE deal_id = deal_id_to_cls;
END;
$$ LANGUAGE 'plpgsql';

CREATE INDEX ad_geos
    ON ad
        USING GIST (geo_position);

CREATE TABLE notifications (
    notification_id bigserial CONSTRAINT notification_pk PRIMARY KEY,
    user_id bigint,
    CONSTRAINT notification_user FOREIGN KEY (user_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    notification_type citext,
    creation_datetime TIMESTAMP WITH TIME ZONE default now(),
    payload bytea,
    is_read boolean NOT NULL DEFAULT false
);

CREATE TABLE comment (
    comment_id bigserial CONSTRAINT comment_pk PRIMARY KEY,
    ad_id bigint,
    CONSTRAINT comment_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    text citext,
    creation_datetime TIMESTAMP WITH TIME ZONE default now(),
    author_id bigint,
    CONSTRAINT comment_user FOREIGN KEY (author_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION
);

CREATE FUNCTION update_comments_count() RETURNS trigger AS $update_comments_count$
    BEGIN
        IF (TG_OP = 'DELETE') THEN
            UPDATE ad SET comments_count = comments_count - 1 WHERE ad_id = OLD.ad_id;
        ELSEIF (TG_OP = 'INSERT') THEN
            UPDATE ad SET comments_count = comments_count + 1 WHERE ad_id = NEW.ad_id;
        end if;
        RETURN NULL;
    END;
$update_comments_count$ LANGUAGE plpgsql;

CREATE TRIGGER update_comments_count AFTER INSERT OR DELETE ON comment
    FOR EACH ROW EXECUTE PROCEDURE update_comments_count();

CREATE FUNCTION ad_view_create() RETURNS trigger AS $ad_view_create$
    BEGIN
        INSERT INTO ad_view (ad_id, views_count) VALUES (new.ad_id, 0);
        RETURN NULL;
    END;
$ad_view_create$ LANGUAGE plpgsql;

CREATE TRIGGER ad_view_create AFTER INSERT ON ad
    FOR EACH ROW EXECUTE PROCEDURE ad_view_create();


-- DROP TABLE IF EXISTS deal;
-- DROP TABLE IF EXISTS ad_subscribers;
-- DROP TABLE IF EXISTS ad_photos;
-- DROP TABLE IF EXISTS ad;
-- DROP TABLE IF EXISTS users;
-- DROP TABLE IF EXISTS ad_view;
-- DROP TABLE IF EXISTS notifications;
-- DROP TABLE IF EXISTS comment;

-- DROP FUNCTION IF EXISTS make_deal;
-- DROP FUNCTION IF EXISTS close_deal_success;
-- DROP FUNCTION IF EXISTS close_deal_fail_by_subscriber;
-- DROP FUNCTION IF EXISTS close_deal_fail_by_author;

DROP TRIGGER IF EXISTS update_comments_count ON comment;
DROP FUNCTION IF EXISTS update_comments_count;
drop trigger if exists ad_view_create on ad;
drop function if exists ad_view_create;
drop trigger if exists users_stats_create on users;
drop function if exists user_stats_create;
drop trigger if exists update_subscribers_count on ad;
drop function if exists update_subscribers_count;


-- DROP INDEX IF EXISTS ad_geos;
-- DROP INDEX IF EXISTS richest;


CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS postgis;


CREATE TABLE IF NOT EXISTS users (
    vk_id bigint NOT NULL CONSTRAINT user_pk PRIMARY KEY,
    name citext,
    surname citext,
    photo_url text,
    registration_date_time TIMESTAMP WITH TIME ZONE default (now() at time zone 'utc')
);

CREATE TABLE IF NOT EXISTS users_carma (
    user_id bigint,
    CONSTRAINT users_carma_users FOREIGN KEY (user_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    current_carma int NOT NULL default 0,
    frozen_carma int NOT NULL default 0,
    cost_frozen int NOT NULL default 1,
    casback_frozen int NOT NULL default 1,
    last_updated  TIMESTAMP WITH TIME ZONE default (now() at time zone 'utc')
);

CREATE TABLE IF NOT EXISTS users_stats (
    user_stats_id bigserial CONSTRAINT users_stats_pk PRIMARY KEY,
    user_id bigint,
    CONSTRAINT users_stats_users FOREIGN KEY (user_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    total_earned_carma bigint not null default 0,
    total_spent_carma bigint not null default 0,
    total_given_ads bigint not null default 0,
    total_received_ads bigint not null default 0,
    total_aborted_ads bigint not null default 0
);

-- DROP TYPE IF EXISTS feedback;
-- DROP TYPE IF EXISTS ad_status;
-- CREATE TYPE feedback AS ENUM ('ls', 'comments', 'other');
-- CREATE TYPE ad_status AS ENUM ('offer', 'chosen', 'closed', 'aborted');

CREATE TABLE IF NOT EXISTS ad (
    ad_id bigserial CONSTRAINT ad_pk PRIMARY KEY,
    author_id bigint,
    CONSTRAINT ad_author FOREIGN KEY (author_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    header citext,
    text citext,
    region citext,
    district citext,
    ad_type text default 'choice',
--     feedback_type text,
    extra_field citext,
    creation_datetime TIMESTAMP WITH TIME ZONE default (now() at time zone 'utc'),
    lat float,
    long float,
    geo_position geography,
    status text DEFAULT 'offer',
    category citext,    -- mb change for enum of categories too
    comments_count int DEFAULT 0,
    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    ls_enabled boolean default true,
    comments_enabled boolean default true,
    extra_enabled boolean default true,
    subscribers_num int default 0
);

CREATE TABLE IF NOT EXISTS ad_view (
    ad_view_id bigserial CONSTRAINT ad_view_pk PRIMARY KEY ,
    ad_id bigint,
    CONSTRAINT ad_view_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    views_count bigint,
    CONSTRAINT ad_view_unique UNIQUE (ad_id)
);

CREATE TABLE IF NOT EXISTS ad_photos (
    ad_photos_id bigserial CONSTRAINT ad_photos_pk PRIMARY KEY ,
    ad_id bigint,
    CONSTRAINT ad_photos_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    photo_url text
);

CREATE TABLE IF NOT EXISTS ad_subscribers (
    ad_subscribers_id bigserial CONSTRAINT ad_subscribers_pk PRIMARY KEY ,
    ad_id bigint,
    CONSTRAINT ad_subscribers_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    subscriber_id bigint,
    CONSTRAINT ad_subscribers_user FOREIGN KEY (subscriber_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    CONSTRAINT ad_subscriber_unique UNIQUE (ad_id, subscriber_id),
    bid int NOT NULL default 0
);

-- DROP TYPE IF EXISTS deal_status;
-- CREATE TYPE deal_status AS ENUM ('open', 'success');

CREATE TABLE IF NOT EXISTS deal (
    deal_id bigserial CONSTRAINT deal_pk PRIMARY KEY,
    ad_id bigint,
    CONSTRAINT deal_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    subscriber_id bigint,
    status text DEFAULT 'open',
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

CREATE OR REPLACE FUNCTION close_deal_success(deal_id_to_upd INT, price_coeff INT) RETURNS void AS $$
    DECLARE
        _ad_id INT;
        _author_id INT;
        _subscriber_id INT;
        _subscribers_num INT;
        _is_auction BOOLEAN;
        _author_gain INT;
    BEGIN
        -- TODO: add auction
        -- deal with statuses
        UPDATE deal SET status = 'success' WHERE deal_id = deal_id_to_upd RETURNING ad_id INTO _ad_id;
        UPDATE ad SET status = 'closed' WHERE ad_id = _ad_id;

        -- acquiring needed variables
        SELECT author_id, ad_type='auction' FROM ad WHERE ad_id = _ad_id INTO _author_id, _is_auction;
        SELECT subscriber_id FROM deal WHERE deal_id = deal_id_to_upd INTO _subscriber_id;

        -- deal with stats
        UPDATE users_stats SET total_given_ads = total_given_ads + 1 WHERE user_id = _author_id;
        UPDATE users_stats SET total_received_ads = total_received_ads + 1 WHERE user_id = _subscriber_id;

        -- deal with carma stuff
        -- look for performance
        if _is_auction THEN
            -- all the subscribers take their carma back
            UPDATE users_carma SET frozen_carma = frozen_carma - a_s.bid FROM ad_subscribers a_s
                WHERE users_carma.user_id = a_s.subscriber_id and a_s.ad_id = _ad_id;
            -- the chosen subscriber
            SELECT bid FROM ad_subscribers WHERE ad_id = _ad_id AND subscriber_id = _subscriber_id INTO _author_gain;
            UPDATE users_carma SET current_carma = current_carma - _author_gain WHERE user_id = _subscriber_id;
            -- author
            UPDATE users_carma SET current_carma = current_carma + _author_gain WHERE user_id = _author_id;
            UPDATE users_stats SET total_earned_carma = total_earned_carma + _author_gain WHERE user_id = _author_id;
            UPDATE users_stats SET total_spent_carma = total_spent_carma + _author_gain WHERE user_id = _subscriber_id;

        else
            -- all the subscribers
            UPDATE users_carma SET frozen_carma = frozen_carma - (cost_frozen - 1)* price_coeff, cost_frozen = cost_frozen -1
            WHERE user_id IN (SELECT subscriber_id FROM ad_subscribers WHERE ad_id = _ad_id AND subscriber_id != _subscriber_id);

            -- chosen (one) subscriber
            UPDATE users_stats SET total_spent_carma = total_spent_carma + price_coeff * casback_frozen FROM users_carma
                WHERE users_carma.user_id = users_stats.user_id AND users_stats.user_id = _subscriber_id;
            UPDATE users_carma SET frozen_carma = frozen_carma - price_coeff * casback_frozen,
                                   current_carma = current_carma - price_coeff * casback_frozen,
                                   casback_frozen = casback_frozen + 1 WHERE user_id = _subscriber_id;
            -- author, we add carma in amount of (subscribers_num * coeff)
            SELECT COUNT(*) FROM ad_subscribers WHERE ad_id = _ad_id INTO _subscribers_num;
            UPDATE users_carma SET current_carma = current_carma + _subscribers_num * price_coeff WHERE user_id = _author_id;
            UPDATE users_stats SET total_earned_carma = total_earned_carma + _subscribers_num * price_coeff WHERE user_id = _author_id;
        end if;
        -- delete subscribers
        DELETE FROM ad_subscribers WHERE ad_id = _ad_id;
        DELETE FROM notifications WHERE ad_id = _ad_id AND user_id = _subscriber_id AND notification_type='ad_close';
    END;
$$ LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION close_deal_fail_by_author(deal_id_to_cls INT) RETURNS void AS $$
    DECLARE _ad_id INT;
            _is_auction BOOLEAN;
            _subscriber_id INT;
    BEGIN
        _ad_id := (SELECT ad_id FROM deal WHERE deal_id = deal_id_to_cls);
        SELECT ad_type='auction' FROM ad WHERE ad_id = _ad_id INTO _is_auction;
        if _is_auction then
            SELECT subscriber_id FROM deal WHERE deal_id = deal_id_to_cls INTO _subscriber_id;
            UPDATE users_carma SET frozen_carma = frozen_carma - a_s.bid FROM ad_subscribers a_s
                WHERE users_carma.user_id = a_s.subscriber_id and a_s.subscriber_id = _subscriber_id AND a_s.ad_id = _ad_id;
            DELETE FROM ad_subscribers WHERE ad_id = _ad_id AND subscriber_id = _subscriber_id;
        end if;
        UPDATE ad SET status = 'offer' WHERE ad_id = _ad_id;
        DELETE FROM deal WHERE deal_id = deal_id_to_cls;
        DELETE FROM notifications WHERE ad_id = _ad_id AND user_id = _subscriber_id AND notification_type='ad_close';
    END;
$$ LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION close_deal_fail_by_subscriber(deal_id_to_cls INT, price_coeff INT) RETURNS void AS $$
DECLARE
    _ad_id INT;
    _author_id INT;
    _is_auction BOOLEAN;
    _subscriber_id INT;
BEGIN
    -- TODO: add auction
    _ad_id := (SELECT ad_id FROM deal WHERE deal_id = deal_id_to_cls);
    UPDATE ad SET status = 'aborted' WHERE ad_id = _ad_id;
    SELECT subscriber_id FROM deal WHERE deal_id = deal_id_to_cls INTO _subscriber_id;
    DELETE FROM deal WHERE deal_id = deal_id_to_cls;

    SELECT ad_type='auction' FROM ad WHERE ad_id = _ad_id INTO _is_auction;
    IF _is_auction THEN
        UPDATE users_carma SET frozen_carma = frozen_carma - a_s.bid FROM ad_subscribers a_s
            WHERE users_carma.user_id = a_s.subscriber_id and a_s.subscriber_id = _subscriber_id AND a_s.ad_id = _ad_id;
        DELETE FROM ad_subscribers WHERE ad_id = _ad_id AND subscriber_id = _subscriber_id;
    else
        UPDATE users_carma SET cost_frozen = cost_frozen -1, frozen_carma = frozen_carma - (cost_frozen - 1) * price_coeff
            WHERE user_id IN (SELECT subscriber_id FROM ad_subscribers WHERE ad_id = _ad_id);
    end if;
    -- look for performance

    DELETE FROM ad_subscribers WHERE ad_id = _ad_id;
    SELECT author_id FROM ad WHERE ad_id = _ad_id INTO _author_id;
    UPDATE users_stats SET total_aborted_ads = total_aborted_ads + 1 WHERE user_id = _author_id;

    DELETE FROM notifications WHERE ad_id = _ad_id AND user_id = _subscriber_id AND notification_type='ad_close';
END;
$$ LANGUAGE 'plpgsql';

CREATE INDEX IF NOT EXISTS ad_geos
    ON ad
        USING GIST (geo_position);

CREATE TABLE IF NOT EXISTS notifications (
    notification_id bigserial CONSTRAINT notification_pk PRIMARY KEY,
    user_id bigint,
    CONSTRAINT notification_user FOREIGN KEY (user_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    ad_id bigint,
    notification_type citext,
    creation_datetime TIMESTAMP WITH TIME ZONE default (now() at time zone 'utc'),
    payload bytea,
    is_read boolean NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS comment (
    comment_id bigserial CONSTRAINT comment_pk PRIMARY KEY,
    ad_id bigint,
    CONSTRAINT comment_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    text citext,
    creation_datetime TIMESTAMP WITH TIME ZONE default (now() at time zone 'utc'),
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

CREATE FUNCTION update_subscribers_count() RETURNS trigger AS $update_subscribers_count$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE ad SET subscribers_count = subscribers_count - 1 WHERE ad_id = OLD.ad_id;
    ELSEIF (TG_OP = 'INSERT') THEN
        UPDATE ad SET subscribers_count = subscribers_count + 1 WHERE ad_id = NEW.ad_id;
    end if;
    RETURN NULL;
END;
$update_subscribers_count$ LANGUAGE plpgsql;

CREATE TRIGGER update_subscribers_count AFTER INSERT OR DELETE ON ad_subscribers
    FOR EACH ROW EXECUTE PROCEDURE update_subscribers_count();

CREATE FUNCTION ad_view_create() RETURNS trigger AS $ad_view_create$
    BEGIN
        INSERT INTO ad_view (ad_id, views_count) VALUES (new.ad_id, 0);
        RETURN NULL;
    END;
$ad_view_create$ LANGUAGE plpgsql;

CREATE TRIGGER ad_view_create AFTER INSERT ON ad
    FOR EACH ROW EXECUTE PROCEDURE ad_view_create();


CREATE FUNCTION user_stats_create() RETURNS trigger AS $user_stats_create$
BEGIN
    INSERT INTO users_stats (user_id) VALUES (new.vk_id);
    INSERT INTO users_carma (user_id) VALUES (new.vk_id);
    RETURN NULL;
END;
$user_stats_create$ LANGUAGE plpgsql;

CREATE TRIGGER users_stats_create AFTER INSERT ON users
    FOR EACH ROW EXECUTE PROCEDURE user_stats_create();

CREATE INDEX IF NOT EXISTS richest ON ad_subscribers (bid);

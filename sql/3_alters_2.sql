ALTER TABLE users add column if not exists last_change_time TIMESTAMP WITH TIME ZONE default (now() at time zone 'utc');

CREATE TABLE IF NOT EXISTS subscribe_history (
    ad_subscribers_id bigserial CONSTRAINT subscribe_history_pk PRIMARY KEY ,
    ad_id bigint,
    CONSTRAINT subscribe_history_ad FOREIGN KEY (ad_id)
        REFERENCES ad (ad_id) ON UPDATE CASCADE ON DELETE CASCADE,
    subscriber_id bigint,
    CONSTRAINT subscribe_history_user FOREIGN KEY (subscriber_id)
        REFERENCES users (vk_id) ON UPDATE CASCADE ON DELETE NO ACTION,
    CONSTRAINT subscribe_history_unique UNIQUE (ad_id, subscriber_id),
    times int NOT NULL default 0
);
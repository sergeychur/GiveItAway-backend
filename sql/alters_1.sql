ALTER TABLE ad DROP column if exists is_auction;
ALTER TABLE ad ADD column if not exists ad_type text default 'choice';
ALTER TABLE ad DROP column if exists feedback_type;
alter table ad add column if not exists ls_enabled boolean default true,
               add column if not exists comments_enabled boolean default true,
               add column if not exists extra_enabled boolean default true,
               add column if not exists subscribers_count int default 0;
alter table notifications add column if not exists ad_id bigint;

alter table ad add column if not exists fts tsvector;
create index if not exists ad_fts on ad using gin(fts);
drop trigger if exists tsvectorupdate on ad;
drop function if exists ad_trigger;
create function ad_trigger() returns trigger as $$
begin
    new.fts := setweight(to_tsvector('ru', coalesce(new.header,'')),
                         'A') ||
               setweight(to_tsvector('ru', coalesce(new.text,'')),
                         'B');
    return new;
end;
$$ LANGUAGE plpgsql;

CREATE TRIGGER  tsvectorupdate before insert or update
    on ad for each row execute procedure ad_trigger();

alter table ad add column if not exists metro citext,
    add column if not exists full_adress citext;

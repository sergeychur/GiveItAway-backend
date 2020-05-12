CREATE TEXT SEARCH DICTIONARY russian_ispell (
    TEMPLATE = ispell,
    DictFile = russian,
    AffFile = russian,
    StopWords = russian
);

CREATE TEXT SEARCH CONFIGURATION ru (COPY=russian);

ALTER TEXT SEARCH CONFIGURATION ru
    ALTER MAPPING FOR hword, hword_part, word
    WITH russian_ispell, russian_stem;

SET default_text_search_config = 'ru';
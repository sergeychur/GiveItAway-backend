FROM serega753/postgis:12
RUN apt-get update
RUN apt-get install -y myspell-ru
RUN cd /usr/share/postgresql/12/tsearch_data
RUN DICT=/usr/share/hunspell/ru_RU
RUN iconv -f koi8-r -t utf-8 -o russian.affix /usr/share/hunspell/ru_RU.aff
RUN iconv -f koi8-r -t utf-8 -o russian.dict  /usr/share/hunspell/ru_RU.dic

ENV POSTGRES_USER=docker \
POSTGRES_DB=docker \
POSTGRES_PASSWORD=docker

#COPY ./sql/full_text_search.sql /docker-entrypoint-initdb.d/
#COPY ./sql/schema.sql /docker-entrypoint-initdb.d/
#todo: check if * works
COPY ./sql/*.sql /docker-entrypoint-initdb.d/

EXPOSE 5432

CMD ["postgres"]
FROM kartoza/postgis

ENV POSTGRES_USER=docker \
POSTGRES_DB=docker \
POSTGRES_PASSWORD=docker

COPY ./sql/schema.sql /docker-entrypoint-initdb.d/

EXPOSE 5432

CMD ["postgres"]
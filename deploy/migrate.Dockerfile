FROM migrate/migrate

WORKDIR /migrations
COPY migrations /migrations

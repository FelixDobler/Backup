#!/bin/bash

docker compose -f docker-compose.yaml up -d

docker compose -f docker-compose.yaml exec nextcloud bash -c "ln -s /usr/bin/echo /usr/local/bin/php"
docker compose -f docker-compose.yaml exec postgres bash -c "ln -s /usr/bin/echo /usr/local/bin/pg_dump"
# ln -s /usr/bin/echo /usr/local/bin/rsync

echo "Mock Nextcloud environment started."

echo "---------------------------------------"
app

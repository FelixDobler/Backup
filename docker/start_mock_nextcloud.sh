#!/bin/bash
# ---------------------------------
# Nextcloud
docker compose -f docker-compose.yaml up -d

docker compose -f docker-compose.yaml exec nextcloud bash -c "ln -s /usr/bin/echo /usr/local/bin/php"
docker compose -f docker-compose.yaml exec postgres bash -c "ln -s /usr/bin/echo /usr/local/bin/pg_dump"
# ln -s /usr/bin/echo /usr/local/bin/rsync

echo "Mock Nextcloud environment started."

# mkdir -p /backups/tmp/nextcloud
# chmod 777 /backups/tmp/nextcloud

# ---------------------------------
# Rsync
mkdir -p /backupDirOnRemote
mkdir -p /rsyncSourceDir/subdir
touch /rsyncSourceDir/testfile.txt
touch /rsyncSourceDir/subdir/testfile2.txt


echo "---------------------------------------"
app

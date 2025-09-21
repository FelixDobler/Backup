#!/bin/bash
set -euo pipefail
echo "$component Backup"
echo $component
echo $rsyncTargetHost
echo $tmpBackupDir
echo $dockerMountDir
# echo $script
echo $dstRelativePath

# send email alert if any of the following commands fail containing the error line and then exit the script
trap 'printf "Subject: BACKUP ALERT\n\nError during nextcloud backup\n$(date)" | msmtp $alert_email; exit 1' ERR

mkdir -p $tmpBackupDir

# enable maintenance mode for data integrity
docker exec -u www-data nextcloud-app-1 php occ maintenance:mode --on

# copy nextcloud data to live dir
rsync -Aax --delete --mkpath $dockerMountDir/nextcloud/ ${rsyncTargetHost}$dstRelativePath/nextcloud/

# create database backup and copy it to live dir
date=$(date +"%Y-%m-%d")
docker exec nextcloud-db-1 pg_dump nextcloud -h localhost -U nextcloud > $tmpBackupDir/nextcloud-sqlbkp-$date.bak

# disable maintenance mode
docker exec -u www-data nextcloud-app-1 php occ maintenance:mode --off

rsync -Aax --delete --mkpath $tmpBackupDir/nextcloud-sqlbkp-$date.bak ${rsyncTargetHost}$dstRelativePath/postgres/nextcloud-sqlbkp.bak

rm $tmpBackupDir/nextcloud-sqlbkp-$date.bak

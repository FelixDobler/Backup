#!/bin/bash
set -euo pipefail
set -x # for debugging
echo "$component Backup"

echo "$json_properties"

echo "Rsync Target Host: $rsyncTargetHost"
tmpBackupDir=$(component_property ".tmpBackupDir")
dockerMountDir=$(component_property ".dockerMountDir")
dstRelativePath=$(component_property ".dstRelativePath")

echo "Temporary Backup Directory: $tmpBackupDir"
echo "Docker Mount Directory: $dockerMountDir"
echo "Destination Relative Path: $dstRelativePath"

# send email alert if any of the following commands fail containing the error line and then exit the script
trap 'printf "Subject: BACKUP ALERT\n\nError during nextcloud backup\n$(date)" | msmtp $alert_email; exit 1' ERR

mkdir -p $tmpBackupDir

# enable maintenance mode for data integrity
docker exec -u www-data nextcloud-app-1 php occ maintenance:mode --on

# copy nextcloud data to live dir
rsync -ax --delete --mkpath $dockerMountDir/nextcloud/ ${rsyncTargetHost}$dstRelativePath/nextcloud/

# create database backup and copy it to live dir
date=$(date +"%Y-%m-%d")
docker exec nextcloud-db-1 pg_dump nextcloud -h localhost -U nextcloud > $tmpBackupDir/nextcloud-sqlbkp-$date.bak

# disable maintenance mode
docker exec -u www-data nextcloud-app-1 php occ maintenance:mode --off

rsync -ax --delete --mkpath $tmpBackupDir/nextcloud-sqlbkp-$date.bak ${rsyncTargetHost}$dstRelativePath/postgres/nextcloud-sqlbkp.bak

rm $tmpBackupDir/nextcloud-sqlbkp-$date.bak

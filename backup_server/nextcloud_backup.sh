#!/bin/bash

# load the config file
configPath="/etc/FDBackup/config.json"
read_config () {
    jq -r "$1" "$configPath"
}

# load alert email address and the location to save the backup to
alert_email=$(read_config ".email")

# load the nextcloud backup directory and main directory from the config file
# for usage with rsync remote host, set it in config using normal rsync format
nextcloudBackupDir=$(read_config ".nextcloud.backupDir")
nextcloudMainDir=$(read_config ".nextcloud.dockerMountDir")

# send email alert if any of the following commands fail containing the error line and then exit the script
trap 'printf "Subject: BACKUP ALERT\n\nError during nextcloud backup\n$(date)" | msmtp $alert_email; exit 1' ERR

# enable maintenance mode for data integrity
docker exec -u www-data nextcloud-app-1 php occ maintenance:mode --on

# create database backup
pgTempDir=$(mktemp -d)
docker exec nextcloud-db-1 pg_dump nextcloud -h localhost -U nextcloud > $pgTempDir/nextcloud-sqlbkp.bak

# copy sql backup and nextcloud data to live dir on target host
# target structure on host
# - rrsyncRootLocation
#   |- nextcloud-sqlbkp.bak
#   |- nextcloud
#      |- AUTHORS
#      |- ....
rsync -aAx --delete --verbose --stats $pgTempDir/nextcloud-sqlbkp.bak $nextcloudMainDir/nextcloud $nextcloudBackupDir/

# cleanup tmp dir
rm -r $pgTempDir

# disable maintenance mode
docker exec -u www-data nextcloud-app-1 php occ maintenance:mode --off

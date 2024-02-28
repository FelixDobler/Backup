#!/bin/bash

# load the config file
configPath="/etc/FDBackup/config.json"
read_config () {
    jq -r "$1" "$configPath"
}

# load alert email address and the location to save the backup to
alert_email=$(read_config ".email")
syncDir=$(read_config ".syncDir")
backupDir=$(read_config ".backupDir")

# load the nextcloud backup directory and main directory from the config file
nextcloudBackupDir=$(read_config ".nextcloud.backupDir")
nextcloudMainDir=$(read_config ".nextcloud.dockerMountDir")

# send email alert if any of the following commands fail containing the error line and then exit the script
trap 'printf "Subject: BACKUP ALERT\n\nError during nextcloud backup\n$(date)" | msmtp $alert_email; exit 1' ERR

# enable maintenance mode for data integrity
docker exec -u www-data nextcloud-app-1 php occ maintenance:mode --on

# copy nextcloud data to live dir
rsync -Aax --delete $nextcloudMainDir/nextcloud/ $nextcloudBackupDir/nextcloud/

# create database backup and copy it to live dir
date=$(date +"%Y-%m-%d")
docker exec nextcloud-db-1 pg_dump nextcloud -h localhost -U nextcloud -f nextcloud-sqlbkp_${date}.bak
docker cp nextcloud-db-1:/nextcloud-sqlbkp_${date}.bak $nextcloudBackupDir/postgres/
        # indicate completion of backup to the alert manager
        touch $backupDir/odroidm1_lastBackup
docker exec nextcloud-db-1 rm nextcloud-sqlbkp_${date}.bak

# disable maintenance mode
docker exec -u www-data nextcloud-app-1 php occ maintenance:mode --off

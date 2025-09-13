#!/bin/bash
# sync 
error=0
{
    # two separate commands to sync the different directory lists: generic linux and host specific
    rsync -ar -v --delete --files-from=file-list-host / backup-server:
    error=$(($error || $?))

    rsync -ar -v --delete --files-from=file-list-linux --delete-excluded --exclude-from=file-exclude / backup-server:
    error=$(($error || $?))
} >backup.log 2>backup.err

# a MTA has to be setup on your machine for the alert to work
[ $error != 0 ] && mail -s "${HOSTNAME} Backup failed" TARGET_MAIL_ADDRESS < backup.err

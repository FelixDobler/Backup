#!/bin/bash

# this file lives in `~/` of the dedicated sync hub user
# and accepts the client rsync connection for their specific directory, given by the clientName

configPath="/etc/FDBackup/config.json"
read_config () {
    jq -r "$1" "$configPath"
}

client_name=$1
backup_dir=$(read_config ".backupDir")
sync_dir=$(read_config ".syncDir")
log_dir=$(read_config ".logDirIncoming")
alert_email=$(read_config ".email")


# redirect errors to logfile
exec 2>$log_dir/$(date -Iminutes)_$client_name.err

# currently not possible because the user has no email set up, maybe use shared config with root or system wide?
    # trap "printf 'Backup Receive Error\n\nError receiving sync from client $client_name\n$(date)' | msmtp $alert_email; exit 1" ERR
trap "echo 'Error receiving sync from client' >&2 ; exit 1" ERR

/usr/bin/rrsync -wo $sync_dir/$client_name

# modification date shows the timestamp of the last client sync
touch $backup_dir/${client_name}_lastBackup

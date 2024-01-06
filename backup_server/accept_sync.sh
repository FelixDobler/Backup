#!/bin/sh
# this file lives in `~/` of the dedicated sync hub user
# and accepts the client rsync connection for their specific directory, given by the hostname
hostname=$1
backup_dir=/backups
sync_dir=$backup_dir/live

/usr/bin/rrsync -wo $sync_dir/$hostname

# modification date shows the timestamp of the last client sync
touch $backup_dir/${hostname}_lastBackup

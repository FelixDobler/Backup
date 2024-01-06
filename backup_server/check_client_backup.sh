#!/bin/bash

backup_dir=/backups
alert_email=felix@doblerones.de

# alert threshold in seconds (e.g., 7 days = 604800 seconds)
alert_threshold=$((86400*1))
# TODO add threshold per client in separate json file

# the file modification date shows the timestamp of the last client sync

for file in $backup_dir/*_lastBackup; do
    hostname=$(basename "$file" "_lastBackup")

    # Get the last modified timestamp
    last_modified=$(stat -c %Y "$file")

    # Get the current timestamp
    current_time=$(date +%s)

    # Calculate the time difference
    time_diff=$((current_time - last_modified))

    # Compare with the threshold
    if [ "$time_diff" -gt "$alert_threshold" ]; then
        relative_elapsed=$(date -ud "@$time_diff" +"$(( $time_diff/3600/24 )) days %H hours")
        printf "Subject: BACKUP ALERT $hostname\n\n\"$hostname\" did not backup for $relative_elapsed" |
        tee /dev/tty | msmtp $alert_email
    fi
done


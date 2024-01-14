#!/bin/bash

read_config () {
    jq -r $1 config.json
}


path_to_remote_repo=$(read_config ".remote.borgPath")
alert_email=$(read_config ".email")

# Get the timestamp of last backup
last_backup_timestr=$(borg list --last 1 --format "{time}" $path_to_remote_repo)
status_borg=$?
last_modified=$(date -d"$last_backup_timestr" "+%s")

# Get the current timestamp
current_time=$(date +%s)

# Calculate the time difference
time_diff=$((current_time - last_modified))

# Get threshold value for client from config
alert_threshold_str="$(read_config ".remote.alertMissingBackupThreshold")"

# convert duration from json to number of seconds
alert_threshold=$(date -d"1970-01-01 00:00:00 UTC $alert_threshold_str" "+%s")

send_alert=0
# Compare with the threshold
if [ "$time_diff" -gt "$alert_threshold" ]; then
    relative_time_elapsed=$(date -ud "@$time_diff" +"$(( $time_diff/3600/24 )) days %H hours")

    alert_msg_header="Subject: Missing Main Backup"
    alert_msg="Main backup server did not backup for $relative_time_elapsed\n"
    send_alert=$(( $send_alert + 1 ))
fi

if [ $status_borg -ne 0 ]; then
    alert_msg_header="Subject: Failed to reach remote Server"
    alert_msg="Failed to acquire last backup time from remote server. There might be a connection issue."
    send_alert=$(( $send_alert + 1 ))
fi

if [ $send_alert -gt 0 ]; then
    printf "$alert_msg_header\n\n$alert_msg" | msmtp $alert_email
    echo "Sent alert"
fi
echo "Check finished"

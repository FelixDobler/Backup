#!/bin/bash

read_config () {
    jq -r $1 config.json
}

backup_dir=$(read_config ".backupDir")
alert_email=$(read_config ".email")

# the file modification date of the `HOSTNAME_lastBackup` shows the timestamp of the last client sync and can be used to detect problems with the backups
send_alert=0
alert_msg_header="Subject: BACKUP ALERT"

# TODO iterate over all hostname in config.json instead
for file in $backup_dir/*_lastBackup; do
    hostname=$(basename "$file" "_lastBackup")

    # Get the last modified timestamp
    last_modified=$(stat -c %Y "$file")

    # Get the current timestamp
    current_time=$(date +%s)

    # Calculate the time difference
    time_diff=$((current_time - last_modified))

    # Get threshold value for client from config
    alert_threshold_str="$(read_config ".clients.$hostname.alertThresholdAge")"

    # convert duration from json to number of seconds
    alert_threshold=$(date -d"1970-01-01 00:00:00 UTC $alert_threshold_str" "+%s")

    # Compare with the threshold
    if [ "$time_diff" -gt "$alert_threshold" ]; then
        relative_time_elapsed=$(date -ud "@$time_diff" +"$(( $time_diff/3600/24 )) days %H hours")

        # change flag to enable email alert
        send_alert=$(( $send_alert + 1 ))
        alert_msg="$alert_msg\"$hostname\" did not backup for $relative_time_elapsed\n"
        echo "Prepare alert: \"$hostname\" did not backup for $relative_time_elapsed"
    fi
done

if [ $send_alert -gt 0 ]; then
    printf "$alert_msg_header\n\n$alert_msg" | msmtp $alert_email
    echo "Sent alert"
fi

echo "Check finished"

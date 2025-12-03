#!/bin/bash
set -euo pipefail
set -x
# echo "$component Backup"
# echo $component
source=$(component_property ".source")
echo $rsyncTargetHost

# echo $files
# echo $excludeFile
dstRelativePath=$(component_property ".dstRelativePath")

hostname=$(hostname)
pwd
component_property ".additionalFiles[]"

exit

# send email alert if any of the following commands fail containing the error line and then exit the script
# trap 'printf "Subject: BACKUP ALERT\n\nError during host backup\n$(date)" | msmtp $alert_email; exit 1' ERR

rsync -ax -v --delete --delete-excluded \
--dry-run \
$( $(component_property -e '.useDefaultFiles') && printf %s "--files-from=files_linux.txt" ) \
$( $(component_property -e '.useDefaultExcludes') && printf %s "--exclude-from=excludes_linux.txt") \

$source ${rsyncTargetHost}$dstRelativePath

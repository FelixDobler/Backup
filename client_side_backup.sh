#!/bin/bash
error=0

{

rsync -ar -v --delete --files-from=file-list-host / backup-server:
error=$(($error || $?))

rsync -ar -v --delete --files-from=file-list-linux --delete-excluded --exclude-from=file-exclude / backup-server:
error=$(($error || $?))

} > backup.log 2>backup.err

[ $error != 0 ] && mail -s "Backup failed" felix@doblerones.de < backup.err

$time = get-date -f yyyy-MM-ddTHH-mm-ss
New-BurntToastNotification -Text "starting Backup" -Silent -UniqueIdentifier "Backup-script"
wsl -d Debian --user root --cd "PATH\TO\local_config" rsync -ar -v --delete --delete-excluded -e "ssh -p 2201 -i /home/felix/.ssh/backup" --files-from=files --exclude-from=exclude / external-backups@192.168.178.91:/ 2>&1 | Tee-Object "PATH\TO\local_config\logs\$time.log"
New-BurntToastNotification -Text "finished Backup" -UniqueIdentifier "Backup-script" -ActivatedAction { explorer.exe  "PATH\TO\local_config\logs\$time.log" }

# Backup
project about my backup plan

- use rsync with rrsync restrictions on the backup server to push data
- use borg backup to archive the live directory locally and on a remote server with encryption

# Rsync
- create user for backups
- backup location on drive with `live` subdir
- add ssh key for every client to remote server with forced command calling `accept_client_backup.sh`
  ```
  command="~/execute_backup.sh HOSTNAME",restrict ssh-ed25519 PUBKEY COMMENT
  ```
- use cronjob on the client for daily backup

# Borg
- create local repo
- add ssh key to remote server with forced `borg serve` command
  ```
  command="borg serve --restrict-to-path PATH_TO_BORG_REPO",restrict ssh-ed25519 PUBKEY COMMENT
  ```
- add remote to ssh config
- create remote repo
- backup both encryption keys
- use `borg_backup.sh` for daily execution on the server, for each local and remote

# Clients
## Windows
- Use WSL Version 1 for acceptable performance on host os file systems
```
sudo rsync -ar -v --delete --delete-excluded -e "ssh -p 2201 -i /home/felix/.ssh/backup" --files-from=files --exclude-from=exclude / external-backups@192.168.178.91:/
```

# Alerts
configure email alerts

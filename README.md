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

# Alerts
configure email alerts

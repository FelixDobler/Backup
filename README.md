# Backup
project about my backup plan

- use rsync with rrsync restrictions on the backup server to push data
- use borg backup to archive the live directory locally and on a remote server with encryption

# Rsync
- create user for backups
- add ssh key to remote server with forced `rrsync` command
- use cronjob on the client for daily backup

# Borg
- create local repo
- add ssh key to remote server with forced `borg serve` command
- create remote repo
- backup both encryption keys
- use script from documentation with cronjob for daily execution on the server, for each local and remote

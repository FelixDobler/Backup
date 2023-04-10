# Backup
documentation of personal backup strategy

- Clients user *rsync* to push their data into a *live directory*. Each client has only access to their own files. \
  This should be fast because rsync only transfers changes in files.
- These files get used to create backups with *borg* that are no acessible by the clients anymore. \
  The Borg backups are deduplicated and compressed.
- One backup is on the same machine as the live directory and one is at a remote place.
- Note that the borg-backups are encrypted, so the data stays protected and the remote-server can't access it.

# Process
-  clients push their data to the live-dir daily
-  after that the backup server creates the local- and remote borg-backup


# Clients

## Linux
- schedule execution of [client_backup](client_side_backup.sh)

## Windows
- Use WSL Version 1 for acceptable performance on host os file systems
- copy and adjust [backup.ps1](local_config/backup.ps1)
- change ownership to `Administrators` and remove write permission for normal users
- Create task in `Task Scheduler` from template
- Install Powershell module [BurntToast](https://github.com/Windos/BurntToast)


# TODO Alerts
configure email alerts
https://wiki.archlinux.org/title/Msmtp

# Remote Backup Server
- start is allowed from 03:30 on
- no rate-limiting required

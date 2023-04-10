- Make an user for the clients to upload their backups (e.g. `external-backups`)
- Install `borg` and `rsync`

# Storage
- prepare the filesystem used for the backups (e.g. setup user permissions, automount, ...)
-  create folders for all the clients you want to backup

# Live Directory
This is used by the clients to upload their current data. \
They only have access to their own files.

Borg will later use these files to create an incremental, deduplicated and compressed backup to which the clients have no access.

# SSH
- create `~/.ssh/authorized_keys` for backup user
- create `~/.ssh/config` for easier handling of remote server
- use *forced command* so only [accept_client_backup.sh](accept_client_backup.sh) is allowed, which restricts access to the folder of the current client
  ```
  command="~/execute_backup.sh HOSTNAME",restrict ssh-ed25519 PUBKEY COMMENT
  ```

# Borg
- Initialize local and remote repository using `keyfile` as encryption method
- ⚠️Backup the encryption key ⚠️
- setup daily cron job for `root` user to take a "snapshot" of the data in the *Live-Directory* for **both local and remote repo**
  
   Use [borg_backup.sh](borg_backup.sh)

# TODO
- prevent changes to *live-dir* for duration of borg backup(s)

  maybe exit early in [client_side_backup.sh](client_side_backup.sh)
  

- Email alert when
  - files from clients are not synced regularly
  - local/remote borg backup fails

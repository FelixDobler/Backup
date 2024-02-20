- Make an user for the clients to upload their backups (e.g. `external-backups`)
- Install `borg` and `rsync`

# Storage
- prepare the filesystem used for the backups (e.g. setup user permissions, automount, ...)
- create folders for all the clients you want to backup

# Sync Hub
Consists of one directory for each client where they to upload their current data. \
They only have access to their own files.

Borg will later use these files to create an incremental, deduplicated and compressed backup to which the clients have no access.

# SSH
- create `~/.ssh/authorized_keys` for backup user
- create `~/.ssh/config` for easier handling of remote server
- use *forced command* so only [the server side sync script](accept_sync.sh) is allowed, which restricts access to the folder of the current client
  ```
  command="~/accept_sync.sh HOSTNAME",restrict ssh-ed25519 PUBKEY COMMENT
  ```

# Borg
- Initialize local and remote repository using `keyfile` as encryption method
- **⚠️Store the encryption key at a safe place⚠️**
- setup daily cron job for `root` user to [create a backup](borg_backup.sh) of the data in the *Sync-Directory* for **both local and remote repo**

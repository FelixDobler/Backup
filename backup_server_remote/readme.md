- Make an user for the backup tasks (e.g. `felix-backups`)
- Install `borg`

# Storage
- create partition and filesystem
- create mountpoint directory and make it immutable so it can't be written to unless the drive is mounted
- add entry in `/etc/fstab` to mount usb drive (add `nofail` to prevent boot issues when drive not present)
- mount the drive using `sudo mount -a`
- `chown` the drive for the backup user

# SSH
- create `~/.ssh/authorized_keys`
- use *forced command* so only borg is allowed to run
  ```
  command="borg serve --restrict-to-path PATH_TO_BORG_REPO",restrict ssh-ed25519 PUBKEY COMMENT
  ```

# Borg
- Initialize repository from first backup server using `keyfile` as encryption method
- **⚠️Store the encryption key at a safe place⚠️**

# TODO
- Email alert when no daily backup is created by the main backup server

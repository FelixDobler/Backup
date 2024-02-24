#!/bin/sh
# script is modified from the official borg documentation

# Paths of the available repositories (path needs to be provided in command):
# local /backups/borg-repo
# remote backup:/mnt/felix-backup/borg-repo

configPath=/etc/FDBackup/config.json
read_config () {
    jq -r "$1" "$configPath"
}

alert_email=$(read_config ".email")
export BORG_REPO=$1

# See the section "Passphrase notes" for more infos.
export BORG_PASSPHRASE=''

# some helpers and error handling:
info() { printf "\n%s %s\n\n" "$( date )" "$*" >&2; }
trap 'echo $( date ) Backup interrupted >&2; exit 2' INT TERM

info "Starting backup"

# Backup the most important directories into an archive named after
# the machine this script is currently running on:

borg create                         \
    --verbose                       \
    --filter AME                    \
    --list                          \
    --stats                         \
    --show-rc                       \
    --compression lz4               \
    --exclude-caches                \
    --exclude 'home/*/.cache/*'     \
    --exclude 'var/tmp/*'           \
                                    \
    ::'combined-{now}'              \
    $syncDir

backup_exit=$?

info "Pruning repository"

# Use the `prune` subcommand to maintain 7 daily, 4 weekly and 6 monthly archives.

borg prune                          \
    --list                          \
    --glob-archives 'combined-*'    \
    --show-rc                       \
    --keep-daily    7               \
    --keep-weekly   4               \
    --keep-monthly  6

prune_exit=$?

# use highest exit code as global exit code
global_exit=$(( backup_exit > prune_exit ? backup_exit : prune_exit ))

if [ ${global_exit} -eq 0 ]; then
    info "Backup and Prune finished successfully"
elif [ ${global_exit} -eq 1 ]; then
    info "Backup and/or Prune finished with warnings"
    printf "Subject: Backup Warning\n\nBackup and/or Prune finished with warnings:\n\tRepository: $1\n\tExit Code: $global_exit" | msmtp $alert_email
else
    info "Backup and/or Prune finished with errors"
    printf "Subject: Backup Error\n\nBackup and/or Prune finished with errors:\n\tRepository: $1\n\tExit Code: $global_exit" | msmtp $alert_email
fi

exit ${global_exit}

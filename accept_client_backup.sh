#!/bin/sh
hostname=$1
live_dir=/backups/live
archived_dir=/backups/archived

/usr/bin/rrsync -wo /backups/live/$hostname
touch /backups/${hostname}_backup

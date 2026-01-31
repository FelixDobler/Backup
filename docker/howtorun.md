```
docker build . --tag backup-go
```

```
docker run --rm --env-file .env -v /var/run/docker.sock:/var/run/docker.sock -v ./go_docker_data:/backupDirOnRemote backup-go
```

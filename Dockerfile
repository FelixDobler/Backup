FROM debian:latest

WORKDIR /usr/src/app

# Add Docker's official GPG key:
RUN apt-get update
RUN apt-get install -y ca-certificates curl
RUN install -m 0755 -d /etc/apt/keyrings
RUN curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
RUN chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
RUN tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/debian
Suites: $(. /etc/os-release && echo "$VERSION_CODENAME")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF

RUN apt update

RUN apt-get install -y golang \
rsync \
docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

RUN mkdir -p /nextcloud-data/nextcloud
RUN touch /nextcloud-data/nextcloud/dummyfile.txt

RUN mkdir -p /backupDirOnRemote

COPY . .
COPY docker .
RUN go build -v -o /usr/local/bin/app cmd/client/main.go

RUN chmod +x start_mock_nextcloud.sh
CMD ["./start_mock_nextcloud.sh"]

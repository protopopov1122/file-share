FROM alpine:latest

# Install dependencies
RUN apk update
RUN apk add go nginx supervisor openssh

# Create directory structure
RUN mkdir -p /file-share/build
RUN mkdir -p /file-share/dist
RUN mkdir -p /file-share/storage
RUN mkdir -p /file-share/user
RUN mkdir -p /file-share/config/persistent/keys
RUN mkdir -p /file-share/config/runtime

# Build project
WORKDIR /file-share/build
COPY src ./src
RUN cd ./src/fileshare && go mod download
RUN cd ./src/fileshare && go build -o /file-share/dist/file-share ./main
RUN cd ./src/fileshare && go test ./lib
RUN cd ./src/uploader && go mod download
RUN cd ./src/uploader && go build -o /file-share/dist/uploader ./main
RUN rm -rf /file-share/build
COPY src/startup.sh /file-share/dist

# Copy configuration
COPY config/api.json /file-share/config/persistent
ADD config/nginx.conf /etc/nginx/nginx.conf
ADD config/supervisord.ini /etc/supervisor.d/supervisord.ini
ADD config/sshd_config /etc/ssh/sshd_config

# Configure users
RUN addgroup -S www -g 1000 && adduser -S www -G www -u 1000
RUN chown -R www:www /file-share/storage
RUN addgroup -S upload && adduser -S upload -D -h /file-share/user -G upload -s /bin/ash
RUN echo "upload:$(cat /dev/urandom | tr -dc '_A-Za-z0-9' | head -c${1:-32})" | chpasswd

# Configure SSH
RUN ssh-keygen -A
COPY config/keys/* /file-share/config/persistent/keys

WORKDIR /file-share/dist
ENTRYPOINT ["/file-share/dist/startup.sh"]

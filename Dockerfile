FROM alpine:latest

# Install dependencies
RUN apk update
RUN apk add go nginx supervisor

# Create directory structure
RUN mkdir -p /file-share/build
RUN mkdir -p /file-share/dist
RUN mkdir -p /file-share/storage

# Build project
WORKDIR /file-share/build
COPY src ./src
RUN cd ./src/fileshare && go mod download
RUN cd ./src/fileshare && go build -o /file-share/dist/file-share ./main
RUN rm -rf /file-share/build
COPY src/startup.sh /file-share/dist

# Copy configuration
COPY config/api.json /file-share/dist
ADD config/nginx.conf /etc/nginx/nginx.conf
ADD config/supervisord.ini /etc/supervisor.d/supervisord.ini

# Configure user
RUN addgroup -S www -g 1000 && adduser -S www -G www -u 1000
RUN chown -R www:www /file-share/storage

WORKDIR /file-share/dist
ENTRYPOINT ["/file-share/dist/startup.sh"]

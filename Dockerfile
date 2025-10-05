FROM alpine:latest

WORKDIR /app

COPY my-proxy /usr/local/bin/my-proxy

RUN chmod +x /usr/local/bin/my-proxy

ENV GIN_MODE=release

EXPOSE 12312

CMD ["my-proxy", "serve"]

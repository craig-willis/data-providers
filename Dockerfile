FROM alpine:3.5


RUN apk --update add bash curl ca-certificates && rm -rf /var/cache/apk/*

COPY ./build/bin/data-provider-linux-amd64 /data-provider-server
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["server"]

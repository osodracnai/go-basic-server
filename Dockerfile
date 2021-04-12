# build stage
FROM alpine

WORKDIR /opt/app
RUN apk add tzdata
COPY server /usr/local/bin/server
CMD ["server", "server"]

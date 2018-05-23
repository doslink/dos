FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/doslink/dos && \
  (cd dos && make gdos) && \
  cp dos/build/bin/gdos /gdos && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /dos && rm -rf /var/cache/apk/*

EXPOSE 8605
EXPOSE 30605

ENTRYPOINT ["/gdos"]

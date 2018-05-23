# Build Gdos in a stock Go builder container
FROM golang:1.10-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /dos
RUN cd /dos && make gdos

# Pull Gdos into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /dos/build/bin/gdos /usr/local/bin/

EXPOSE 8605 8606 30605 30605/udp
ENTRYPOINT ["gdos"]

FROM ubuntu:xenial

ENV PATH=/usr/lib/go-1.9/bin:$PATH

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang-1.9 git make gcc libc-dev ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/doslink/dos && \
  (cd dos && make gdos) && \
  cp dos/build/bin/gdos /gdos && \
  apt-get remove -y golang-1.9 git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /dos

EXPOSE 8605
EXPOSE 30605

ENTRYPOINT ["/gdos"]

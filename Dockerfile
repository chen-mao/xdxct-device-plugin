FROM ubuntu:20.04 as build

RUN  sed -i s@/archive.ubuntu.com/@/mirror.sjtu.edu.cn/@g /etc/apt/sources.list
RUN  apt-get clean

RUN apt-get update && apt-get install -y --no-install-recommends \
        g++ \
        ca-certificates \
        wget && \
    rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.19.3
RUN wget -nv -O - https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /go/src/xdxct-device-plugin
COPY . .

RUN export CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' && \
    go install -ldflags="-s -w" -v xdxct-device-plugin


FROM debian:stretch-slim

ENV XDXCT_VISIBLE_DEVICES=all
ENV XDXCT_DRIVER_CAPABILITIES=utility

COPY --from=build /go/bin/xdxct-device-plugin /usr/bin/xdxct-device-plugin

CMD ["xdxct-device-plugin"]

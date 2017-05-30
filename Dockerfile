FROM alpine:3.5

ARG V=bad

ADD dist/soil-$V-linux-amd64.tar.gz /usr/bin/
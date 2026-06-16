FROM alpine:latest
ARG TARGETARCH
COPY bin/stockbatch-linux-${TARGETARCH}    /usr/local/bin/stockbatch
COPY bin/filterjson-linux-${TARGETARCH}    /usr/local/bin/filterjson
COPY bin/stockclient-linux-${TARGETARCH}   /usr/local/bin/stockclient
COPY bin/currentreturn-linux-${TARGETARCH} /usr/local/bin/currentreturn
COPY bin/targetreturn-linux-${TARGETARCH}  /usr/local/bin/targetreturn

# FROM alpine:edge

# ENV GOPATH /go
# ENV PATH /go/src/github.com/ssimunic/gossm/bin:$PATH

# ADD . /go/src/github.com/ssimunic/gossm

# RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
#     && apk add --no-cache --update bash ca-certificates \
#     && apk add --no-cache --virtual .build-deps go gcc git libc-dev \
#     && mkdir -p /configs /usr/local/bin /var/log/gossm \
#     && go get github.com/gregdel/pushover \
#     && cd /go/src/github.com/ssimunic/gossm \
#     && go build -v -o /usr/local/bin/gossm cmd/gossm/main.go \
#     && apk del --purge .build-deps \
#     && rm -rf /var/cache/apk*

# ADD configs /configs

# CMD ["gossm", "-config", "/configs/default.json", "-http", ":8080", "-log", "/var/log/gossm/gossm.log"]

# EXPOSE 8080



# Build stage
FROM golang:1.17.1-alpine AS app-builder

ADD . /src

ENV CGO_ENABLED=0
RUN cd /src && \
    ls -la && \
    go build -v -o /main ./cmd/gossm/main.go

# Run stage
FROM alpine:3.13

RUN apk add --no-cache ca-certificates tzdata

COPY --from=app-builder /main /
COPY --from=app-builder /src/configs/default.json /configs/default.json

ENTRYPOINT ["/main"]

# Usage:
#   $ docker build -t shopctl:latest .
#   $ docker run --rm -it -v ~/.config/shopctl:/root/.config/shopctl shopctl

FROM golang:1.24-alpine3.21 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

COPY . .

RUN set -eux; \
    env ; \
    ls -la ; \
    apk add -U --no-cache make git ; \
    make deps install

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/bin/shopctl /bin/shopctl

ENTRYPOINT ["/bin/sh"]

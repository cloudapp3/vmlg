# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=0.1.0-docker
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -buildvcs=false -trimpath \
    -ldflags="-s -w -X github.com/cloudapp3/vmlg/internal/version.Version=${VERSION} -X github.com/cloudapp3/vmlg/internal/version.Commit=${COMMIT} -X github.com/cloudapp3/vmlg/internal/version.BuildTime=${BUILD_TIME}" \
    -o /out/vmlg ./cmd/vmlg

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata iperf3 && \
    addgroup -g 10001 -S vmlg && \
    adduser -u 10001 -S -G vmlg -H -s /sbin/nologin vmlg && \
    mkdir -p /etc/vmlg && \
    chown -R vmlg:vmlg /etc/vmlg

COPY --from=builder /out/vmlg /usr/local/bin/vmlg

USER 10001:10001
WORKDIR /etc/vmlg

EXPOSE 8080 8081 5201-5210/tcp 5201-5210/udp

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/vmlg", "healthcheck"]

ENTRYPOINT ["/usr/local/bin/vmlg"]
CMD []

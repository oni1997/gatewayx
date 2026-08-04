FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o /gatewayx \
    ./apps/gateway

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S gatewayx && adduser -S gatewayx -G gatewayx

COPY --from=builder /gatewayx /usr/local/bin/gatewayx

RUN mkdir -p /etc/gatewayx /var/lib/gatewayx && \
    chown -R gatewayx:gatewayx /etc/gatewayx /var/lib/gatewayx

USER gatewayx

EXPOSE 8080 9090

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["gatewayx"]
CMD ["-c", "/etc/gatewayx/gatewayx.yaml"]

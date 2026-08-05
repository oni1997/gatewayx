FROM golang:1.23-alpine AS builder

ENV GOTOOLCHAIN=auto

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

FROM node:23-alpine AS dashboard-builder

WORKDIR /dashboard
COPY apps/dashboard/package.json apps/dashboard/package-lock.json ./
RUN npm ci
COPY apps/dashboard/ ./
RUN npm run build
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

RUN addgroup -S gatewayx && adduser -S gatewayx -G gatewayx

COPY --from=builder /gatewayx /usr/local/bin/gatewayx
COPY --from=dashboard-builder /dashboard/dist /opt/gatewayx/dashboard/dist

RUN mkdir -p /etc/gatewayx /var/lib/gatewayx && \
    chown -R gatewayx:gatewayx /etc/gatewayx /var/lib/gatewayx /opt/gatewayx

USER gatewayx

ENV GATEWAYX_CONFIG=/etc/gatewayx/gatewayx.yaml

EXPOSE 8080 9090

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["gatewayx"]

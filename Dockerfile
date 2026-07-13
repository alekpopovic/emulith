FROM golang:1.22.5-alpine3.20 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILT=unknown
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.built=${BUILT}" -o /out/emulith ./cmd/emulith

FROM alpine:3.20.3
LABEL org.opencontainers.image.title="Emulith" \
      org.opencontainers.image.description="Open-source local cloud emulator for development and CI" \
      org.opencontainers.image.licenses="AGPL-3.0-or-later" \
      org.opencontainers.image.source="https://github.com/alekpopovic/emulith"
RUN apk add --no-cache ca-certificates && addgroup -S -g 10001 emulith && adduser -S -D -H -u 10001 -G emulith emulith && mkdir -p /var/lib/emulith && chown emulith:emulith /var/lib/emulith
COPY --from=builder /out/emulith /usr/local/bin/emulith
COPY LICENSE NOTICE /licenses/
USER 10001:10001
VOLUME ["/var/lib/emulith"]
EXPOSE 4566
ENTRYPOINT ["/usr/local/bin/emulith"]
CMD ["serve", "--addr", ":4566", "--data-dir", "/var/lib/emulith"]

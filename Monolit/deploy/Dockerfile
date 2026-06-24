FROM golang:1.25.7-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/calllens ./cmd

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl ffmpeg tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --system calllens \
    && useradd --system --gid calllens --home-dir /app --shell /usr/sbin/nologin calllens

WORKDIR /app

COPY --from=builder /out/calllens /app/calllens
COPY migrations /app/migrations

RUN mkdir -p /app/uploads \
    && chown -R calllens:calllens /app

USER calllens

EXPOSE 8080

CMD ["/app/calllens"]

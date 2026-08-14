FROM golang:1.26.5-trixie AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -o /out/recap ./cmd/app \
    && CGO_ENABLED=0 go build -trimpath -o /out/seed ./cmd/seed

FROM debian:13-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system app \
    && useradd --system --gid app --no-create-home app

WORKDIR /app

COPY --from=builder --chown=app:app /out/recap ./recap
COPY --from=builder --chown=app:app /out/seed ./seed

USER app

EXPOSE 8080

ENTRYPOINT ["/app/recap"]

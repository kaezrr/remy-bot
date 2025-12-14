FROM golang:1.25.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /usr/local/bin/remy ./cmd/remy

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /usr/local/bin/remy /usr/local/bin/remy

COPY config.json /app/config.json

RUN mkdir -p data/session && chown -R 1000:1000 /app

USER 1000

ENTRYPOINT ["/usr/local/bin/remy"]

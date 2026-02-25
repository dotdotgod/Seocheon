# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make gcc musl-dev linux-headers

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . .
RUN go build -o /seocheon ./cmd/seocheon

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates jq bash

COPY --from=builder /seocheon /usr/local/bin/seocheon

EXPOSE 26656 26657 1317 9090

ENTRYPOINT ["seocheon"]
CMD ["start"]

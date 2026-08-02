# ---- Build stage ----
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /build/server ./cmd/server

# ---- Run stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /build/server /usr/local/bin/gedis-server

EXPOSE 8080 7000

ENTRYPOINT ["gedis-server"]

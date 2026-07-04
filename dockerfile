# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/go-api ./src

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/go-api /app/go-api
COPY --from=builder /src/templates /app/templates
COPY --from=builder /src/static /app/static

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/go-api"]

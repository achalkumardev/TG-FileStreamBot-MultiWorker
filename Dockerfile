FROM golang:1.25-alpine3.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/fsb -ldflags="-w -s" ./cmd/fsb

FROM scratch
COPY --from=builder /app/fsb /app/fsb
EXPOSE 80
ENTRYPOINT ["/app/fsb", "run"]

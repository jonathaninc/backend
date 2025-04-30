FROM golang:1.24-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o backend ./cmd/main.go

FROM alpine:3.21.2

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/backend .

RUN adduser -D appuser
USER appuser

ENTRYPOINT ["./backend"]

EXPOSE 8080

# Builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOEXPERIMENT=jsonv2 go build -a -installsuffix cgo -o /app/server ./cmd/api/main.go

# Runtime
FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

COPY --from=builder /app/server .

USER appuser

EXPOSE 8080

CMD ["./server"]
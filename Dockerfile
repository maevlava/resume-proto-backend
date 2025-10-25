# Builder
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache build-base mupdf libffi-dev musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOEXPERIMENT=jsonv2 go build -tags "ncgo musl" -o /app/server ./cmd/api/main.go

# Runtime
FROM alpine:latest

RUN apk add --no-cache mupdf libffi

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

COPY --from=builder /app/server .

RUN mkdir -p /home/appuser/uploads && chown -R appuser:appgroup /home/appuser

USER appuser

EXPOSE 8080

CMD ["./server"]
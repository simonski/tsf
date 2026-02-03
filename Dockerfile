FROM golang:1.24-alpine AS builder

RUN apk --no-cache add gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /task ./cmd/task-unified

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

COPY --from=builder /task /app/

EXPOSE 8080

CMD ["/app/task", "server"]

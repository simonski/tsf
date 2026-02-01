FROM golang:1.23-alpine AS builder

RUN apk --no-cache add gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /task-server ./cmd/server
RUN CGO_ENABLED=1 GOOS=linux go build -o /task-initdb ./cmd/initdb
RUN CGO_ENABLED=1 GOOS=linux go build -o /task-orchestrator ./cmd/orchestrator
RUN CGO_ENABLED=1 GOOS=linux go build -o /task-worker ./cmd/worker

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

COPY --from=builder /task-server /app/
COPY --from=builder /task-initdb /app/
COPY --from=builder /task-orchestrator /app/
COPY --from=builder /task-worker /app/

EXPOSE 8080

CMD ["/app/task-server"]

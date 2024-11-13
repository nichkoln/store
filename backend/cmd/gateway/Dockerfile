
FROM golang:1.23.1-alpine AS builder


RUN apk update && apk add --no-cache git


WORKDIR /app

COPY ./backend/go.mod ./backend/go.sum ./


RUN go mod download

COPY . .

WORKDIR /app/backend/cmd/gateway
RUN go build -o /app/api-gateway

FROM alpine:latest

ENV PORT=8000


WORKDIR /root/


COPY --from=builder /app/api-gateway .


EXPOSE 8000

CMD ["./api-gateway"]

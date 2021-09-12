# syntax=docker/dockerfile:1
FROM golang:1.16.8-alpine3.14 AS builder

WORKDIR /app

COPY src ./src/
COPY go.mod ./

RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates
RUN pwd && ls -la .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ddns-updater src/ddns-updater.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/ddns-updater ./

CMD [ "./cloudflare-updater" ]
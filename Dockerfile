FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /app/mobius-hotline-bot . && chmod a+x /app/mobius-hotline-bot

FROM debian:stable-slim AS ca-builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates

RUN update-ca-certificates

FROM scratch

# copy the ca-certificate.crt from the build stage
COPY --from=ca-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/mobius-hotline-bot /app/mobius-hotline-bot

ENTRYPOINT ["/app/mobius-hotline-bot"]

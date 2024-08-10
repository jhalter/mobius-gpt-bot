FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /app/mobius-hotline-bot . && chmod a+x /app/mobius-hotline-bot

FROM debian:stable-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates

RUN update-ca-certificates

COPY --from=builder /app/mobius-hotline-bot /app/mobius-hotline-bot

ENTRYPOINT ["/app/mobius-hotline-bot"]

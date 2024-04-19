FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /app/mobius-hotline-bot . && chmod a+x /app/mobius-hotline-bot

FROM debian:stable-slim

# Change these as you see fit. This makes bind mounting easier so you don't have to edit bind mounted config files as root.
ARG USERNAME=hl-bot
ARG UID=1001
ARG GUID=1001

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates

RUN update-ca-certificates

COPY --from=builder /app/mobius-hotline-bot /app/mobius-hotline-bot

RUN useradd -d /app -u ${UID} ${USERNAME}
RUN chown -R ${USERNAME}:${USERNAME} /app

USER ${USERNAME}
ENTRYPOINT ["/app/mobius-hotline-bot"]

FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-X main.version=$(git describe --exact-match --tags)" -o /app/mobius-hotline-bot . && chmod a+x /app/mobius-hotline-bot

FROM scratch

COPY --from=builder /app/mobius-hotline-bot /app/mobius-hotline-bot

ENTRYPOINT ["/app/mobius-hotline-bot"]

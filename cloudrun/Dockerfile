FROM alpine:latest

WORKDIR /app

COPY ./notifier /app/notifier

RUN chmod +x /app/notifier

ENTRYPOINT ["/app/notifier"]

FROM nats:2.8.1-alpine3.15

RUN apk update && apk add curl

HEALTHCHECK --interval=1s --timeout=1s --start-period=60s CMD ["curl", "-f", "http://localhost:8222/healthz"]

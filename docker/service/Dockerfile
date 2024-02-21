FROM golang:1.21 AS builder

ARG CMD_NAME

WORKDIR /srv/

COPY go.mod /srv/go.mod
COPY go.sum /srv/go.sum
RUN --mount=type=cache,target=/root/.cache/go-build go mod download

COPY cmd /srv/cmd
COPY internal /srv/internal
COPY pkg /srv/pkg

RUN --mount=type=cache,target=/root/.cache/go-build go build -v -o entrypoint /srv/cmd/${CMD_NAME}/main.go

FROM golang:1.21 AS runner

RUN apt-get update && apt-get install -y sqlite3

WORKDIR /srv/

COPY --from=builder /srv/entrypoint /srv/entrypoint

HEALTHCHECK --interval=1s --timeout=1s --start-period=60s CMD ["curl", "-f", "http://localhost:80/healthz"]

STOPSIGNAL SIGINT

# note: ensure to use JSON form (square brackets) otherwise Docker wraps CMD w/ bash (which breaks STOPSIGNALs)
CMD ["/srv/entrypoint"]

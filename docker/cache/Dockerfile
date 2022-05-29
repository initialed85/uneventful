FROM redis:7.0.0-alpine

HEALTHCHECK --interval=1s --timeout=1s --start-period=60s CMD ["redis-cli", "ping"]

CMD ["redis-server", "--save 60 1", "--loglevel warning"]

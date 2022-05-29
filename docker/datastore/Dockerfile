FROM timescale/timescaledb-ha:pg13-ts2.5-latest

HEALTHCHECK --interval=1s --timeout=1s --start-period=60s CMD ["psql", "-U", "postgres", "datastore", "-c", "SELECT 1;"]

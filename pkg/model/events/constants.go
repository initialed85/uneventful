package events

const (
	tableName            = "event"
	createUniqueIndexSQL = "CREATE UNIQUE INDEX IF NOT EXISTS event_id_created_at ON %v (event_id, created_at DESC);"
	createdHypertableSQL = "SELECT create_hypertable('%v', 'created_at', 'event_id', 1, chunk_time_interval => INTERVAL '1 day', if_not_exists => true);"
)

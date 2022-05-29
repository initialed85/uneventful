package states

const (
	tableName            = "state"
	createUniqueIndexSQL = "CREATE UNIQUE INDEX IF NOT EXISTS version_id_created_at ON %v (version_id, created_at DESC);"
	createdHypertableSQL = "SELECT create_hypertable('%v', 'created_at', 'version_id', 1, chunk_time_interval => INTERVAL '1 day', if_not_exists => true);"
)

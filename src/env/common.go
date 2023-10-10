package env

// Environment variable names
const (
	// Port on which REST API will be exposed, defaults to "8000"
	API_PORT = "API_PORT"
	// Address on which REST API will bind. Defaults to "",
	API_BIND_ADDR = "API_BIND_ADDR"

	// Broker fee in tenth of bps
	BROKER_FEE_TBPS = "BROKER_FEE_TBPS"
	// REDIS connection string
	REDIS_ADDR = "REDIS_ADDR"
	REDIS_PW   = "REDIS_PW"
	WS_ADDR    = "WS_ADDR"
	// chainConfig.json configuration file path
	CONFIG_PATH = "CONFIG_PATH"
	// file with private key
	KEYFILE_PATH = "KEYFILE_PATH"
	// Broker key
	BROKER_KEY = "BROKER_KEY"
)

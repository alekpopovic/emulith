package config

import "os"

const (
	DefaultAddr    = ":4566"
	DefaultDataDir = "./data"
)

type Config struct {
	Addr                                            string
	DataDir                                         string
	Region                                          string
	GCPProject                                      string
	GCPPubSubAddr, GCPFirestoreAddr, GCPStorageAddr string
}

func FromEnvironment() Config {
	return Config{
		Addr:             valueOrDefault("EMULITH_ADDR", DefaultAddr),
		DataDir:          valueOrDefault("EMULITH_DATA_DIR", DefaultDataDir),
		Region:           valueOrDefault("EMULITH_AWS_REGION", "us-east-1"),
		GCPProject:       valueOrDefault("EMULITH_GCP_PROJECT_ID", "emulith-local"),
		GCPPubSubAddr:    valueOrDefault("EMULITH_GCP_PUBSUB_ADDR", ":8085"),
		GCPFirestoreAddr: valueOrDefault("EMULITH_GCP_FIRESTORE_ADDR", ":8080"),
		GCPStorageAddr:   valueOrDefault("EMULITH_GCP_STORAGE_ADDR", ":9023"),
	}
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

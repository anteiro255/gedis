package config

const envPrefix = "GEDIS_"

type Config struct {
	Server  *ServerConfig
	Storage *StorageConfig
	Raft    *RaftConfig
}

func Load() (cfg *Config) {
	cfg = &Config{}

	cfg.Server = LoadServerConfig()
	cfg.Storage = LoadStorageConfig()
	cfg.Raft = LoadRaftCfg()
	return
}

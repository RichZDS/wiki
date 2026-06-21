package model

type Config struct {
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	RAG    RAGConfig    `yaml:"rag"`
	Job    JobConfig    `yaml:"job"`
	Env    string       `yaml:"-"`
}

type ServerConfig struct {
	Port           string   `yaml:"port"`
	AppName        string   `yaml:"app_name"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RAGConfig struct {
	IndexName           string `yaml:"index_name"`
	KeyPrefix           string `yaml:"key_prefix"`
	VectorField         string `yaml:"vector_field"`
	VectorDim           int    `yaml:"vector_dim"`
	VectorIndexType     string `yaml:"vector_index_type"`
	BatchSize           int    `yaml:"batch_size"`
	DefaultTopK         int    `yaml:"default_top_k"`
	MaxTopK             int    `yaml:"max_top_k"`
	HNSWMaxEdgesPerNode int    `yaml:"hnsw_max_edges_per_node"`
	HNSWEFConstruction  int    `yaml:"hnsw_ef_construction"`
	HNSWEFRuntime       int    `yaml:"hnsw_ef_runtime"`
}

type JobConfig struct {
	LogDBLevel string `yaml:"log_db_level"`
}

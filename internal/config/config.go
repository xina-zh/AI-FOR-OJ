package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "configs/config.yaml"

const (
	defaultDBCharset   = "utf8mb4"
	defaultDBCollation = "utf8mb4_unicode_ci"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	HTTP     HTTPConfig     `yaml:"http"`
	Log      LogConfig      `yaml:"log"`
	Database DatabaseConfig `yaml:"database"`
	Sandbox  SandboxConfig  `yaml:"sandbox"`
	LLM      LLMConfig      `yaml:"llm"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type HTTPConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type DatabaseConfig struct {
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port"`
	User              string        `yaml:"user"`
	Password          string        `yaml:"password"`
	Name              string        `yaml:"name"`
	Params            string        `yaml:"params"`
	AutoMigrate       bool          `yaml:"auto_migrate"`
	MaxOpenConns      int           `yaml:"max_open_conns"`
	MaxIdleConns      int           `yaml:"max_idle_conns"`
	ConnMaxLifetime   time.Duration `yaml:"conn_max_lifetime"`
	InitMaxRetries    int           `yaml:"init_max_retries"`
	InitRetryInterval time.Duration `yaml:"init_retry_interval"`
}

type SandboxConfig struct {
	WorkDir          string        `yaml:"work_dir"`
	DockerImage      string        `yaml:"docker_image"`
	CompileTimeout   time.Duration `yaml:"compile_timeout"`
	RunTimeoutBuffer time.Duration `yaml:"run_timeout_buffer"`
	CompileMemoryMB  int           `yaml:"compile_memory_mb"`
}

type LLMConfig struct {
	Provider     string        `yaml:"provider"`
	BaseURL      string        `yaml:"base_url"`
	APIKey       string        `yaml:"api_key"`
	Model        string        `yaml:"model"`
	Timeout      time.Duration `yaml:"timeout"`
	MockResponse string        `yaml:"mock_response"`
}

func Load(path string) (Config, error) {
	cfg := defaultConfig()

	resolvedPath, strict := resolveConfigPath(path)
	if resolvedPath != "" {
		data, err := os.ReadFile(resolvedPath)
		if err != nil {
			if strict || !errors.Is(err, os.ErrNotExist) {
				return Config{}, fmt.Errorf("read config file %q: %w", resolvedPath, err)
			}
		} else {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return Config{}, fmt.Errorf("unmarshal config file %q: %w", resolvedPath, err)
			}
		}
	}

	applyEnvOverrides(&cfg)
	return cfg, nil
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.User, c.Password, c.Host, c.Port, c.Name, c.normalizedParams())
}

func (c DatabaseConfig) normalizedParams() string {
	values, err := url.ParseQuery(c.Params)
	if err != nil {
		values = url.Values{}
	}

	values.Set("charset", defaultDBCharset)
	values.Set("collation", defaultDBCollation)

	return values.Encode()
}

func defaultConfig() Config {
	return Config{
		App: AppConfig{
			Name: "ai-for-oj",
			Env:  "local",
		},
		HTTP: HTTPConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    5 * time.Minute,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
		Database: DatabaseConfig{
			Host:              "127.0.0.1",
			Port:              3306,
			User:              "root",
			Password:          "root",
			Name:              "ai_for_oj",
			Params:            "charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local",
			AutoMigrate:       true,
			MaxOpenConns:      20,
			MaxIdleConns:      10,
			ConnMaxLifetime:   30 * time.Minute,
			InitMaxRetries:    10,
			InitRetryInterval: 3 * time.Second,
		},
		Sandbox: SandboxConfig{
			WorkDir:          "/tmp/ai-for-oj-sandbox",
			DockerImage:      "gcc:13",
			CompileTimeout:   10 * time.Second,
			RunTimeoutBuffer: 500 * time.Millisecond,
			CompileMemoryMB:  512,
		},
		LLM: LLMConfig{
			Provider: "mock",
			Model:    "mock-cpp17",
			Timeout:  60 * time.Second,
		},
	}
}

func resolveConfigPath(path string) (string, bool) {
	if envPath := os.Getenv("APP_CONFIG_PATH"); envPath != "" {
		return envPath, true
	}
	if path != "" {
		return path, true
	}
	return defaultConfigPath, false
}

func applyEnvOverrides(cfg *Config) {
	cfg.App.Name = getEnvString("APP_NAME", cfg.App.Name)
	cfg.App.Env = getEnvString("APP_ENV", cfg.App.Env)

	cfg.HTTP.Host = getEnvString("HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = getEnvInt("HTTP_PORT", cfg.HTTP.Port)
	cfg.HTTP.ReadTimeout = getEnvDuration("HTTP_READ_TIMEOUT", cfg.HTTP.ReadTimeout)
	cfg.HTTP.WriteTimeout = getEnvDuration("HTTP_WRITE_TIMEOUT", cfg.HTTP.WriteTimeout)
	cfg.HTTP.IdleTimeout = getEnvDuration("HTTP_IDLE_TIMEOUT", cfg.HTTP.IdleTimeout)
	cfg.HTTP.ShutdownTimeout = getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", cfg.HTTP.ShutdownTimeout)

	cfg.Log.Level = getEnvString("LOG_LEVEL", cfg.Log.Level)
	cfg.Log.Format = getEnvString("LOG_FORMAT", cfg.Log.Format)

	cfg.Database.Host = getEnvString("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnvInt("DB_PORT", cfg.Database.Port)
	cfg.Database.User = getEnvString("DB_USER", cfg.Database.User)
	cfg.Database.Password = getEnvString("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.Name = getEnvString("DB_NAME", cfg.Database.Name)
	cfg.Database.Params = getEnvString("DB_PARAMS", cfg.Database.Params)
	cfg.Database.AutoMigrate = getEnvBool("DB_AUTO_MIGRATE", cfg.Database.AutoMigrate)
	cfg.Database.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", cfg.Database.MaxOpenConns)
	cfg.Database.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", cfg.Database.MaxIdleConns)
	cfg.Database.ConnMaxLifetime = getEnvDuration("DB_CONN_MAX_LIFETIME", cfg.Database.ConnMaxLifetime)
	cfg.Database.InitMaxRetries = getEnvInt("DB_INIT_MAX_RETRIES", cfg.Database.InitMaxRetries)
	cfg.Database.InitRetryInterval = getEnvDuration("DB_INIT_RETRY_INTERVAL", cfg.Database.InitRetryInterval)

	cfg.Sandbox.WorkDir = getEnvString("SANDBOX_WORK_DIR", cfg.Sandbox.WorkDir)
	cfg.Sandbox.DockerImage = getEnvString("SANDBOX_DOCKER_IMAGE", cfg.Sandbox.DockerImage)
	cfg.Sandbox.CompileTimeout = getEnvDuration("SANDBOX_COMPILE_TIMEOUT", cfg.Sandbox.CompileTimeout)
	cfg.Sandbox.RunTimeoutBuffer = getEnvDuration("SANDBOX_RUN_TIMEOUT_BUFFER", cfg.Sandbox.RunTimeoutBuffer)
	cfg.Sandbox.CompileMemoryMB = getEnvInt("SANDBOX_COMPILE_MEMORY_MB", cfg.Sandbox.CompileMemoryMB)

	cfg.LLM.Provider = getEnvString("LLM_PROVIDER", cfg.LLM.Provider)
	cfg.LLM.BaseURL = getEnvString("LLM_BASE_URL", cfg.LLM.BaseURL)
	cfg.LLM.APIKey = getEnvString("LLM_API_KEY", cfg.LLM.APIKey)
	cfg.LLM.Model = getEnvString("LLM_MODEL", cfg.LLM.Model)
	cfg.LLM.Timeout = getEnvDuration("LLM_TIMEOUT", cfg.LLM.Timeout)
	cfg.LLM.MockResponse = getEnvString("LLM_MOCK_RESPONSE", cfg.LLM.MockResponse)
}

func getEnvString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

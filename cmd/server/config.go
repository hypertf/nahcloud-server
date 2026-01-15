package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hypertf/nahcloud/service/chaos"
)

// Version is set at build time
var Version = "dev"

// Config holds all server configuration
type Config struct {
	Addr      string      `mapstructure:"addr"`
	Token     string      `mapstructure:"token"`
	SQLiteDSN string      `mapstructure:"sqlite_dsn"`
	Chaos     ChaosConfig `mapstructure:"chaos"`
}

// ChaosConfig holds chaos engineering configuration
type ChaosConfig struct {
	Enabled      bool              `mapstructure:"enabled"`
	Seed         int64             `mapstructure:"seed"`
	Latency      LatencyConfig     `mapstructure:"latency"`
	ErrorRate    ErrorRateConfig   `mapstructure:"error_rate"`
	ErrorTypes   []int             `mapstructure:"error_types"`
	ErrorWeights []int             `mapstructure:"error_weights"`
}

// LatencyConfig holds latency injection settings
type LatencyConfig struct {
	GlobalMS    string `mapstructure:"global_ms"`
	ProjectsMS  string `mapstructure:"projects_ms"`
	InstancesMS string `mapstructure:"instances_ms"`
	MetadataMS  string `mapstructure:"metadata_ms"`
}

// ErrorRateConfig holds error injection rates
type ErrorRateConfig struct {
	Projects    float64 `mapstructure:"projects"`
	ProjectsGet float64 `mapstructure:"projects_get"`
	Instances   float64 `mapstructure:"instances"`
	Metadata    float64 `mapstructure:"metadata"`
}

// setupConfig initializes viper with flags, env vars, and config file support
func setupConfig(cmd *cobra.Command) {
	// Define flags
	cmd.Flags().StringP("config", "c", "", "Config file path (YAML, JSON, or TOML)")
	cmd.Flags().String("addr", ":8080", "HTTP server address")
	cmd.Flags().String("token", "", "Authentication token")
	cmd.Flags().String("sqlite-dsn", "", "SQLite database path")

	// Chaos flags
	cmd.Flags().Bool("chaos-enabled", false, "Enable chaos engineering")
	cmd.Flags().Int64("chaos-seed", 0, "Random seed for chaos (0 = use current time)")
	cmd.Flags().String("chaos-latency-global", "", "Global latency range in ms (e.g., \"10-100\")")
	cmd.Flags().String("chaos-latency-projects", "", "Projects latency range in ms")
	cmd.Flags().String("chaos-latency-instances", "", "Instances latency range in ms")
	cmd.Flags().String("chaos-latency-metadata", "", "Metadata latency range in ms")
	cmd.Flags().Float64("chaos-errrate-projects", 0.0, "Error rate for projects (0.0-1.0)")
	cmd.Flags().Float64("chaos-errrate-projects-get", 0.0, "Error rate for projects GET (0.0-1.0)")
	cmd.Flags().Float64("chaos-errrate-instances", 0.0, "Error rate for instances (0.0-1.0)")
	cmd.Flags().Float64("chaos-errrate-metadata", 0.0, "Error rate for metadata (0.0-1.0)")
	cmd.Flags().IntSlice("chaos-error-types", []int{503, 500, 429}, "Error HTTP status codes to inject")
	cmd.Flags().IntSlice("chaos-error-weights", []int{3, 2, 1}, "Weights for error types")

	// Bind flags to viper
	viper.BindPFlag("addr", cmd.Flags().Lookup("addr"))
	viper.BindPFlag("token", cmd.Flags().Lookup("token"))
	viper.BindPFlag("sqlite_dsn", cmd.Flags().Lookup("sqlite-dsn"))
	viper.BindPFlag("chaos.enabled", cmd.Flags().Lookup("chaos-enabled"))
	viper.BindPFlag("chaos.seed", cmd.Flags().Lookup("chaos-seed"))
	viper.BindPFlag("chaos.latency.global_ms", cmd.Flags().Lookup("chaos-latency-global"))
	viper.BindPFlag("chaos.latency.projects_ms", cmd.Flags().Lookup("chaos-latency-projects"))
	viper.BindPFlag("chaos.latency.instances_ms", cmd.Flags().Lookup("chaos-latency-instances"))
	viper.BindPFlag("chaos.latency.metadata_ms", cmd.Flags().Lookup("chaos-latency-metadata"))
	viper.BindPFlag("chaos.error_rate.projects", cmd.Flags().Lookup("chaos-errrate-projects"))
	viper.BindPFlag("chaos.error_rate.projects_get", cmd.Flags().Lookup("chaos-errrate-projects-get"))
	viper.BindPFlag("chaos.error_rate.instances", cmd.Flags().Lookup("chaos-errrate-instances"))
	viper.BindPFlag("chaos.error_rate.metadata", cmd.Flags().Lookup("chaos-errrate-metadata"))
	viper.BindPFlag("chaos.error_types", cmd.Flags().Lookup("chaos-error-types"))
	viper.BindPFlag("chaos.error_weights", cmd.Flags().Lookup("chaos-error-weights"))

	// Set up environment variable binding with NAH_ prefix
	viper.SetEnvPrefix("NAH")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("addr", ":8080")
	viper.SetDefault("chaos.error_types", []int{503, 500, 429})
	viper.SetDefault("chaos.error_weights", []int{3, 2, 1})
}

// loadConfig loads configuration from flags, env vars, and config file
func loadConfig(cmd *cobra.Command) (*Config, error) {
	// Check for config file
	configFile, _ := cmd.Flags().GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Handle seed default (use current time if not set)
	if cfg.Chaos.Seed == 0 {
		cfg.Chaos.Seed = time.Now().UnixNano()
	}

	// If projects_get error rate not explicitly set, inherit from projects
	if cfg.Chaos.ErrorRate.ProjectsGet == 0 && cfg.Chaos.ErrorRate.Projects > 0 {
		cfg.Chaos.ErrorRate.ProjectsGet = cfg.Chaos.ErrorRate.Projects
	}

	return &cfg, nil
}

// ToChaosConfig converts our config to the chaos service's Config type
func (c *Config) ToChaosConfig() *chaos.Config {
	cfg := &chaos.Config{
		Enabled:              c.Chaos.Enabled,
		Seed:                 c.Chaos.Seed,
		GlobalLatencyRange:   parseLatencyRange(c.Chaos.Latency.GlobalMS),
		ProjectsLatencyRange: parseLatencyRange(c.Chaos.Latency.ProjectsMS),
		InstancesLatencyRange: parseLatencyRange(c.Chaos.Latency.InstancesMS),
		MetadataLatencyRange: parseLatencyRange(c.Chaos.Latency.MetadataMS),
		ProjectsErrorRate:    c.Chaos.ErrorRate.Projects,
		ProjectsGetErrorRate: c.Chaos.ErrorRate.ProjectsGet,
		InstancesErrorRate:   c.Chaos.ErrorRate.Instances,
		MetadataErrorRate:    c.Chaos.ErrorRate.Metadata,
		ErrorTypes:           c.Chaos.ErrorTypes,
		ErrorWeights:         c.Chaos.ErrorWeights,
	}

	// Use defaults if not set
	if len(cfg.ErrorTypes) == 0 {
		cfg.ErrorTypes = []int{503, 500, 429}
	}
	if len(cfg.ErrorWeights) == 0 {
		cfg.ErrorWeights = []int{3, 2, 1}
	}

	return cfg
}

// parseLatencyRange parses a "min-max" string into a LatencyRange
func parseLatencyRange(value string) *chaos.LatencyRange {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return nil
	}

	min, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	max, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

	if err1 != nil || err2 != nil || min < 0 || max < min {
		return nil
	}

	return &chaos.LatencyRange{Min: min, Max: max}
}

// printConfigHelp prints additional help about environment variables
func printConfigHelp() string {
	return `
Environment Variables:
  All flags can be set via environment variables with the NAH_ prefix.
  Nested keys use underscores. Examples:

  NAH_ADDR=:9090                    Set server address
  NAH_TOKEN=secret                  Set auth token
  NAH_SQLITE_DSN=./data.db          Set database path
  NAH_CHAOS_ENABLED=true            Enable chaos engineering
  NAH_CHAOS_LATENCY_GLOBAL_MS=10-100  Set global latency range

Config File:
  Use --config to specify a YAML, JSON, or TOML config file.
  Example YAML config:

    addr: ":8080"
    token: "secret"
    sqlite_dsn: "./nahcloud.db"
    chaos:
      enabled: true
      seed: 12345
      latency:
        global_ms: "10-100"
      error_rate:
        projects: 0.1
      error_types: [503, 500, 429]
      error_weights: [3, 2, 1]

Priority (highest to lowest):
  1. Command-line flags
  2. Environment variables
  3. Config file
  4. Defaults
`
}

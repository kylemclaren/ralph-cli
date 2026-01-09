package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all Ralph configuration
type Config struct {
	Agent         AgentConfig         `mapstructure:"agent"`
	Loop          LoopConfig          `mapstructure:"loop"`
	Paths         PathsConfig         `mapstructure:"paths"`
	Hooks         HooksConfig         `mapstructure:"hooks"`
	Notifications NotificationsConfig `mapstructure:"notifications"`
}

// AgentConfig configures the AI coding agent
type AgentConfig struct {
	Type    string        `mapstructure:"type"`    // claude-code, amp, opencode, codex, custom
	Command string        `mapstructure:"command"` // custom command template
	Flags   []string      `mapstructure:"flags"`   // additional flags
	Timeout time.Duration `mapstructure:"timeout"` // max time per iteration
}

// LoopConfig configures the Ralph loop behavior
type LoopConfig struct {
	MaxIterations      int           `mapstructure:"maxIterations"`
	SleepBetween       time.Duration `mapstructure:"sleepBetween"`
	StopOnFirstFailure bool          `mapstructure:"stopOnFirstFailure"`
}

// PathsConfig configures file paths
type PathsConfig struct {
	PRD      string `mapstructure:"prd"`
	Progress string `mapstructure:"progress"`
	Prompt   string `mapstructure:"prompt"`
}

// HooksConfig configures lifecycle hooks
type HooksConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	OnStart     []string `mapstructure:"onStart"`
	OnIteration []string `mapstructure:"onIteration"`
	OnComplete  []string `mapstructure:"onComplete"`
	OnFailure   []string `mapstructure:"onFailure"`
}

// NotificationsConfig configures notifications
type NotificationsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			Type:    "claude-code",
			Timeout: 30 * time.Minute,
		},
		Loop: LoopConfig{
			MaxIterations:      25,
			SleepBetween:       2 * time.Second,
			StopOnFirstFailure: false,
		},
		Paths: PathsConfig{
			PRD:      ".ralph/prd.json",
			Progress: ".ralph/progress.txt",
			Prompt:   ".ralph/prompt.md",
		},
		Hooks: HooksConfig{
			Enabled: true,
		},
		Notifications: NotificationsConfig{
			Enabled: false,
		},
	}
}

// Load loads configuration from file, environment, and applies defaults
func Load(cfgFile string) (*Config, error) {
	// Set defaults
	setDefaults()

	// Set up environment variable binding
	viper.SetEnvPrefix("RALPH")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Try to find and read config file
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for ralph.yaml in current directory and .ralph/
		viper.SetConfigName("ralph")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath(".ralph")
	}

	// Read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	// Unmarshal into config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	defaults := DefaultConfig()

	viper.SetDefault("agent.type", defaults.Agent.Type)
	viper.SetDefault("agent.timeout", defaults.Agent.Timeout)
	viper.SetDefault("loop.maxIterations", defaults.Loop.MaxIterations)
	viper.SetDefault("loop.sleepBetween", defaults.Loop.SleepBetween)
	viper.SetDefault("loop.stopOnFirstFailure", defaults.Loop.StopOnFirstFailure)
	viper.SetDefault("paths.prd", defaults.Paths.PRD)
	viper.SetDefault("paths.progress", defaults.Paths.Progress)
	viper.SetDefault("paths.prompt", defaults.Paths.Prompt)
	viper.SetDefault("hooks.enabled", defaults.Hooks.Enabled)
	viper.SetDefault("notifications.enabled", defaults.Notifications.Enabled)
}

// GetAgentCommand returns the full command for the configured agent
func (c *Config) GetAgentCommand() (string, []string, error) {
	switch c.Agent.Type {
	case "claude-code":
		return "claude", append([]string{"--dangerously-skip-permissions"}, c.Agent.Flags...), nil
	case "amp":
		return "amp", append([]string{"--dangerously-allow-all"}, c.Agent.Flags...), nil
	case "opencode":
		return "opencode", c.Agent.Flags, nil
	case "codex":
		return "codex", c.Agent.Flags, nil
	case "custom":
		if c.Agent.Command == "" {
			return "", nil, fmt.Errorf("custom agent type requires agent.command to be set")
		}
		// Parse custom command
		parts := strings.Fields(c.Agent.Command)
		if len(parts) == 0 {
			return "", nil, fmt.Errorf("invalid custom command")
		}
		return parts[0], append(parts[1:], c.Agent.Flags...), nil
	default:
		return "", nil, fmt.Errorf("unknown agent type: %s", c.Agent.Type)
	}
}

// EnsureDirectories creates necessary directories for Ralph files
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		filepath.Dir(c.Paths.PRD),
		filepath.Dir(c.Paths.Progress),
		filepath.Dir(c.Paths.Prompt),
	}

	for _, dir := range dirs {
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}
	return nil
}

// ConfigFileUsed returns the config file that was loaded
func ConfigFileUsed() string {
	return viper.ConfigFileUsed()
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config provides access to CLI configuration.
type Config interface {
	// BaseURL returns the APISIX Admin API base URL for the active context.
	BaseURL() string
	// APIKey returns the API key for the active context.
	APIKey() string
	// CurrentContext returns the name of the active context.
	CurrentContext() string
	// Contexts returns all configured contexts.
	Contexts() []Context
	// GetContext returns a context by name.
	GetContext(name string) (*Context, error)
	// AddContext adds a new context. Returns error if name already exists.
	AddContext(ctx Context) error
	// RemoveContext removes a context by name.
	RemoveContext(name string) error
	// SetCurrentContext sets the active context by name.
	SetCurrentContext(name string) error
	// Save persists the configuration to disk.
	Save() error
}

// Context represents a connection to an APISIX instance.
type Context struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
	APIKey string `yaml:"api-key,omitempty"`
}

// fileData is the on-disk YAML structure.
type fileData struct {
	CurrentContext string    `yaml:"current-context"`
	Contexts       []Context `yaml:"contexts"`
}

// FileConfig implements Config by reading/writing a YAML config file.
type FileConfig struct {
	mu       sync.RWMutex
	path     string
	data     fileData
	loaded   bool
	override configOverride
}

// configOverride holds flag/env overrides that take precedence over file config.
type configOverride struct {
	server string
	apiKey string
}

// NewFileConfig creates a FileConfig that reads from the default config path.
// It does not read the file until the first access (lazy loading).
func NewFileConfig() *FileConfig {
	return &FileConfig{
		path: defaultConfigPath(),
	}
}

// NewFileConfigWithPath creates a FileConfig with a specific file path.
func NewFileConfigWithPath(path string) *FileConfig {
	return &FileConfig{
		path: path,
	}
}

// SetServerOverride sets a server URL that takes precedence over the config file.
func (c *FileConfig) SetServerOverride(server string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.override.server = server
}

// SetAPIKeyOverride sets an API key that takes precedence over the config file.
func (c *FileConfig) SetAPIKeyOverride(apiKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.override.apiKey = apiKey
}

// BaseURL returns the APISIX Admin API base URL.
// Precedence: override (flag/env) > active context config.
func (c *FileConfig) BaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.override.server != "" {
		return c.override.server
	}

	ctx := c.currentCtx()
	if ctx != nil {
		return ctx.Server
	}
	return ""
}

// APIKey returns the API key.
// Precedence: override (flag/env) > active context config.
func (c *FileConfig) APIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.override.apiKey != "" {
		return c.override.apiKey
	}

	ctx := c.currentCtx()
	if ctx != nil {
		return ctx.APIKey
	}
	return ""
}

// CurrentContext returns the name of the active context.
func (c *FileConfig) CurrentContext() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_ = c.ensureLoaded()
	return c.data.CurrentContext
}

// Contexts returns all configured contexts.
func (c *FileConfig) Contexts() []Context {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_ = c.ensureLoaded()

	out := make([]Context, len(c.data.Contexts))
	copy(out, c.data.Contexts)
	return out
}

// GetContext returns a context by name.
func (c *FileConfig) GetContext(name string) (*Context, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if err := c.ensureLoaded(); err != nil {
		return nil, err
	}

	for i := range c.data.Contexts {
		if c.data.Contexts[i].Name == name {
			ctx := c.data.Contexts[i]
			return &ctx, nil
		}
	}
	return nil, fmt.Errorf("context %q not found", name)
}

// AddContext adds a new context. Returns error if name already exists.
func (c *FileConfig) AddContext(ctx Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureLoaded(); err != nil {
		return err
	}

	for _, existing := range c.data.Contexts {
		if existing.Name == ctx.Name {
			return fmt.Errorf("context %q already exists", ctx.Name)
		}
	}

	c.data.Contexts = append(c.data.Contexts, ctx)

	// Auto-set as current if it's the first context.
	if c.data.CurrentContext == "" {
		c.data.CurrentContext = ctx.Name
	}
	return nil
}

// RemoveContext removes a context by name.
func (c *FileConfig) RemoveContext(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureLoaded(); err != nil {
		return err
	}

	idx := -1
	for i, ctx := range c.data.Contexts {
		if ctx.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("context %q not found", name)
	}

	c.data.Contexts = append(c.data.Contexts[:idx], c.data.Contexts[idx+1:]...)

	// Clear current context if we just removed it.
	if c.data.CurrentContext == name {
		c.data.CurrentContext = ""
		if len(c.data.Contexts) > 0 {
			c.data.CurrentContext = c.data.Contexts[0].Name
		}
	}
	return nil
}

// SetCurrentContext sets the active context by name.
func (c *FileConfig) SetCurrentContext(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureLoaded(); err != nil {
		return err
	}

	found := false
	for _, ctx := range c.data.Contexts {
		if ctx.Name == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("context %q not found", name)
	}

	c.data.CurrentContext = name
	return nil
}

// Save writes the configuration to disk.
func (c *FileConfig) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	b, err := yaml.Marshal(&c.data)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.path, b, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// Path returns the config file path (for display/debugging).
func (c *FileConfig) Path() string {
	return c.path
}

// ensureLoaded lazily loads the config file. Must be called with at least a read lock held.
// If the file doesn't exist, it initializes with empty defaults.
func (c *FileConfig) ensureLoaded() error {
	if c.loaded {
		return nil
	}

	b, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			c.data = fileData{}
			c.loaded = true
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(b, &c.data); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	c.loaded = true
	return nil
}

// currentCtx returns the active context or nil. Must be called with at least a read lock held.
func (c *FileConfig) currentCtx() *Context {
	_ = c.ensureLoaded()
	for i := range c.data.Contexts {
		if c.data.Contexts[i].Name == c.data.CurrentContext {
			return &c.data.Contexts[i]
		}
	}
	return nil
}

// defaultConfigPath returns the default config file path.
// Precedence: A6_CONFIG_DIR > XDG_CONFIG_HOME > ~/.config
func defaultConfigPath() string {
	if dir := os.Getenv("A6_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "config.yaml")
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "a6", "config.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "a6", "config.yaml")
}

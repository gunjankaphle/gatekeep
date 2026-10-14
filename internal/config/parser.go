package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser handles parsing YAML configuration files
type Parser struct {
	validator *Validator
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{
		validator: NewValidator(),
	}
}

// ParseFile parses a YAML configuration file
func (p *Parser) ParseFile(path string) (*Config, error) {
	// Read file
	// #nosec G304 - File path is user-provided config file, this is intentional
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses YAML configuration from bytes
func (p *Parser) Parse(data []byte) (*Config, error) {
	var config Config

	// Fail closed: ignored desired fields could otherwise become revocations.
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, fmt.Errorf("failed to parse trailing YAML: %w", err)
		}
		return nil, fmt.Errorf("configuration must contain exactly one YAML document")
	}

	// Validate configuration
	if err := p.validator.Validate(&config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &config, nil
}

// ParseWithoutValidation parses YAML without validation (useful for testing)
func (p *Parser) ParseWithoutValidation(data []byte) (*Config, error) {
	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

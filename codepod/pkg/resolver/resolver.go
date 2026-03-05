// Package resolver provides URI resolution for CodePod hybrid URI scheme.
// It supports both plain text identities (for simple providers like wsl, ssh-remote)
// and Hex-encoded JSON identities (for complex providers like dev-container).
package resolver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

// Provider types supported by CodePod
const (
	ProviderWSL            = "wsl"
	ProviderSSHRemote       = "ssh-remote"
	ProviderDockerContainer = "docker-container"
	ProviderDevContainer    = "dev-container"
)

// Authority represents a parsed URI authority component.
type Authority struct {
	Provider string      // Provider type (wsl, ssh-remote, docker-container, dev-container)
	Identity string      // Identity (plain text or Hex-encoded JSON)
	IdentityParsed any   // Parsed identity (string or *DevContainerConfig)
	Original   string   // Original authority string
}

// DevContainerConfig represents the configuration for a dev-container.
type DevContainerConfig struct {
	Path   string         `json:"path"`   // Project path in the container
	Config string         `json:"config"` // Relative path to .devcontainer.json
	Hash   string         `json:"hash"`   // Config hash for idempotency
	Env    map[string]string `json:"env"` // Environment variables to inject
}

// SplitAuthorityAndPath splits a URI string into authority and path components.
// Returns the authority part and the path part (including leading slash).
func SplitAuthorityAndPath(uri string) (authority, path string, err error) {
	uri = strings.TrimPrefix(uri, "codepod-remote://")

	// Find the first slash that separates authority from path
	slashIdx := strings.Index(uri, "/")

	if slashIdx == -1 {
		// No path, only authority
		return uri, "", nil
	}

	if slashIdx == 0 {
		// Path starts immediately (no authority)
		return "", uri[1:], nil
	}

	return uri[:slashIdx], uri[slashIdx:], nil
}

// Resolve parses an authority string and returns an Authority struct.
// It automatically detects whether to use plain or Hex-JSON parsing based on provider type.
func Resolve(authority string) (*Authority, error) {
	if authority == "" {
		return nil, errors.New("missing authority")
	}

	// Split provider from identity by '+' delimiter
	plusIdx := strings.Index(authority, "+")
	if plusIdx == -1 {
		return nil, errors.New("missing identity in authority")
	}

	provider := authority[:plusIdx]
	identity := authority[plusIdx+1:]

	if provider == "" {
		return nil, errors.New("missing provider in authority")
	}

	if identity == "" {
		return nil, errors.New("missing identity in authority")
	}

	auth := &Authority{
		Provider: provider,
		Identity: identity,
		Original: authority,
	}

	// Parse identity based on provider type
	switch provider {
	case ProviderWSL, ProviderSSHRemote, ProviderDockerContainer:
		// Plain identity
		auth.IdentityParsed = identity
	case ProviderDevContainer:
		// Hex-JSON identity
		config, err := DecodeHexJSON(identity)
		if err != nil {
			return nil, err
		}
		auth.IdentityParsed = config
	default:
		return nil, errors.New("unknown provider: " + provider)
	}

	return auth, nil
}

// EncodeToHex encodes a JSON object to Hex string.
func EncodeToHex(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// DecodeFromHex decodes a Hex string back to JSON bytes.
func DecodeFromHex(hexStr string) ([]byte, error) {
	// Validate hex string
	if len(hexStr)%2 != 0 {
		return nil, errors.New("invalid hex: odd length")
	}

	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, errors.New("invalid hex: non-hex character")
	}

	return data, nil
}

// DecodeHexJSON decodes a Hex-encoded JSON string to DevContainerConfig.
func DecodeHexJSON(hexStr string) (*DevContainerConfig, error) {
	data, err := DecodeFromHex(hexStr)
	if err != nil {
		return nil, err
	}

	var config DevContainerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.New("invalid json in hex: " + err.Error())
	}

	return &config, nil
}

// EncodeHexJSON encodes a DevContainerConfig to Hex string.
func EncodeHexJSON(config *DevContainerConfig) (string, error) {
	return EncodeToHex(config)
}

// IdentityParser defines the interface for parsing identity based on provider type.
type IdentityParser interface {
	Parse(identity string) (any, error)
}

// PlainIdentityParser parses plain text identities.
type PlainIdentityParser struct{}

// Parse returns the identity as-is (plain string).
func (p *PlainIdentityParser) Parse(identity string) (any, error) {
	return identity, nil
}

// HexJSONIdentityParser parses Hex-encoded JSON identities.
type HexJSONIdentityParser struct{}

// Parse decodes Hex string to DevContainerConfig.
func (p *HexJSONIdentityParser) Parse(identity string) (any, error) {
	return DecodeHexJSON(identity)
}

// GetIdentityParser returns the appropriate identity parser for a given provider.
func GetIdentityParser(provider string) IdentityParser {
	switch provider {
	case ProviderWSL, ProviderSSHRemote, ProviderDockerContainer:
		return &PlainIdentityParser{}
	case ProviderDevContainer:
		return &HexJSONIdentityParser{}
	default:
		return nil
	}
}

// Resolver handles URI resolution with hybrid identity parsing.
type Resolver struct {
	// Can be extended with custom providers
}

// NewResolver creates a new Resolver instance.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve is a method on Resolver for consistency.
func (r *Resolver) Resolve(authority string) (*Authority, error) {
	return Resolve(authority)
}

package resolver

import (
	"testing"
)

func TestSplitAuthorityAndPath(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		wantAuth   string
		wantPath   string
		wantErr    bool
	}{
		{
			name:     "with path",
			uri:      "wsl+ubuntu/home/user",
			wantAuth: "wsl+ubuntu",
			wantPath: "/home/user",
		},
		{
			name:     "without path",
			uri:      "docker-container+nginx-dev",
			wantAuth: "docker-container+nginx-dev",
			wantPath: "",
		},
		{
			name:     "with trailing slash",
			uri:      "ssh-remote+host/",
			wantAuth: "ssh-remote+host",
			wantPath: "/",
		},
		{
			name:     "with multiple path segments",
			uri:      "dev-container+abc123/a/b/c/d",
			wantAuth: "dev-container+abc123",
			wantPath: "/a/b/c/d",
		},
		{
			name:     "preserve encoded characters",
			uri:      "wsl+ubuntu/path%20with%20spaces/file",
			wantAuth: "wsl+ubuntu",
			wantPath: "/path%20with%20spaces/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAuth, gotPath, err := SplitAuthorityAndPath(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitAuthorityAndPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAuth != tt.wantAuth {
				t.Errorf("SplitAuthorityAndPath() gotAuth = %v, want %v", gotAuth, tt.wantAuth)
			}
			if gotPath != tt.wantPath {
				t.Errorf("SplitAuthorityAndPath() gotPath = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name       string
		authority  string
		wantProv   string
		wantIdent  string
		wantErr    bool
	}{
		{
			name:      "WSL provider",
			authority: "wsl+ubuntu",
			wantProv:  "wsl",
			wantIdent: "ubuntu",
		},
		{
			name:      "SSH remote provider",
			authority: "ssh-remote+192.168.1.1",
			wantProv:  "ssh-remote",
			wantIdent: "192.168.1.1",
		},
		{
			name:      "Docker container provider",
			authority: "docker-container+nginx-dev",
			wantProv:  "docker-container",
			wantIdent: "nginx-dev",
		},
		{
			name:      "empty authority",
			authority: "",
			wantErr:   true,
		},
		{
			name:      "missing provider",
			authority: "+ubuntu",
			wantErr:   true,
		},
		{
			name:      "missing identity",
			authority: "wsl",
			wantErr:   true,
		},
		{
			name:      "unknown provider",
			authority: "unknown+test",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.authority)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Provider != tt.wantProv {
				t.Errorf("Resolve() got.Provider = %v, want %v", got.Provider, tt.wantProv)
			}
			if got.Identity != tt.wantIdent {
				t.Errorf("Resolve() got.Identity = %v, want %v", got.Identity, tt.wantIdent)
			}
		})
	}
}

func TestDevContainerResolve(t *testing.T) {
	// Test dev-container with Hex-JSON
	// {"config":".devcontainer.json","hash":"a1b2c3d4","path":"/home/user"}
	hexStr := "7b22636f6e666967223a222e646576636f6e7461696e65722e6a736f6e222c2268617368223a226131623263336434222c2270617468223a222f686f6d652f75736572227d"
	authority := "dev-container+" + hexStr

	got, err := Resolve(authority)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if got.Provider != "dev-container" {
		t.Errorf("Resolve() got.Provider = %v, want dev-container", got.Provider)
	}

	config, ok := got.IdentityParsed.(*DevContainerConfig)
	if !ok {
		t.Fatalf("Resolve() got.IdentityParsed = %T, want *DevContainerConfig", got.IdentityParsed)
	}

	if config.Path != "/home/user" {
		t.Errorf("Resolve() config.Path = %v, want /home/user", config.Path)
	}
	if config.Config != ".devcontainer.json" {
		t.Errorf("Resolve() config.Config = %v, want .devcontainer.json", config.Config)
	}
	if config.Hash != "a1b2c3d4" {
		t.Errorf("Resolve() config.Hash = %v, want a1b2c3d4", config.Hash)
	}
}

func TestEncodeDecodeHex(t *testing.T) {
	// Test round-trip encode/decode
	original := map[string]string{
		"path":   "/home/user",
		"config": ".devcontainer.json",
	}

	encoded, err := EncodeToHex(original)
	if err != nil {
		t.Fatalf("EncodeToHex() error = %v", err)
	}

	decoded, err := DecodeFromHex(encoded)
	if err != nil {
		t.Fatalf("DecodeFromHex() error = %v", err)
	}

	if string(decoded) != `{"config":".devcontainer.json","path":"/home/user"}` {
		t.Errorf("Round-trip failed: got %s", string(decoded))
	}
}

func TestDevContainerConfig(t *testing.T) {
	config := &DevContainerConfig{
		Path:   "/abs/path/to/project",
		Config: ".devcontainer.json",
		Hash:   "a1b2c3d4",
		Env:    map[string]string{"DEBUG": "true"},
	}

	encoded, err := EncodeHexJSON(config)
	if err != nil {
		t.Fatalf("EncodeHexJSON() error = %v", err)
	}

	decoded, err := DecodeHexJSON(encoded)
	if err != nil {
		t.Fatalf("DecodeHexJSON() error = %v", err)
	}

	if decoded.Path != config.Path {
		t.Errorf("Path mismatch: got %v, want %v", decoded.Path, config.Path)
	}
	if decoded.Config != config.Config {
		t.Errorf("Config mismatch: got %v, want %v", decoded.Config, config.Config)
	}
	if decoded.Hash != config.Hash {
		t.Errorf("Hash mismatch: got %v, want %v", decoded.Hash, config.Hash)
	}
	if decoded.Env["DEBUG"] != "true" {
		t.Errorf("Env mismatch: got %v, want true", decoded.Env["DEBUG"])
	}
}

func TestInvalidHex(t *testing.T) {
	tests := []struct {
		name      string
		hexStr    string
		wantErr   bool
		errContains string
	}{
		{
			name:        "odd length",
			hexStr:      "abc",
			wantErr:     true,
			errContains: "odd length",
		},
		{
			name:        "non-hex character",
			hexStr:      "7b22gg",
			wantErr:     true,
			errContains: "non-hex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeFromHex(tt.hexStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeFromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("DecodeFromHex() error = %v, want containing %v", err, tt.errContains)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

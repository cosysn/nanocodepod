package workspace

import (
	"testing"
)

func TestResourceLimits_Validate(t *testing.T) {
	tests := []struct {
		name    string
		limits  *ResourceLimits
		wantErr bool
	}{
		{
			name:    "valid CPU",
			limits:  &ResourceLimits{CPU: "1.0"},
			wantErr: false,
		},
		{
			name:    "valid CPU 2",
			limits:  &ResourceLimits{CPU: "2"},
			wantErr: false,
		},
		{
			name:    "invalid CPU",
			limits:  &ResourceLimits{CPU: "abc"},
			wantErr: true,
		},
		{
			name:    "zero CPU",
			limits:  &ResourceLimits{CPU: "0"},
			wantErr: false,
		},
		{
			name:    "empty",
			limits:  &ResourceLimits{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.limits.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ResourceLimits.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceLimits_ToDockerResources(t *testing.T) {
	limits := &ResourceLimits{
		CPU:    "2.0",
		Memory: "1g",
		CPUset: "0,1",
	}

	resources := limits.ToDockerResources()

	if resources["nano-cpus"] != "2.0" {
		t.Errorf("expected nano-cpus 2.0, got %s", resources["nano-cpus"])
	}
	if resources["memory"] != "1g" {
		t.Errorf("expected memory 1g, got %s", resources["memory"])
	}
	if resources["cpuset-cpus"] != "0,1" {
		t.Errorf("expected cpuset-cpus 0,1, got %s", resources["cpuset-cpus"])
	}
}

func TestDefaultResources(t *testing.T) {
	resources := DefaultResources()

	if resources.CPU != "1.0" {
		t.Errorf("expected CPU 1.0, got %s", resources.CPU)
	}
	if resources.Memory != "512m" {
		t.Errorf("expected Memory 512m, got %s", resources.Memory)
	}
}

func TestGetResourcesFromConfig(t *testing.T) {
	settings := map[string]string{
		"cpu":    "4",
		"memory": "2g",
		"cpuset": "0",
	}

	resources := GetResourcesFromConfig(settings)

	if resources.CPU != "4" {
		t.Errorf("expected CPU 4, got %s", resources.CPU)
	}
	if resources.Memory != "2g" {
		t.Errorf("expected Memory 2g, got %s", resources.Memory)
	}
	if resources.CPUset != "0" {
		t.Errorf("expected CPUset 0, got %s", resources.CPUset)
	}
}

func TestGetResourcesFromConfig_Nil(t *testing.T) {
	resources := GetResourcesFromConfig(nil)

	if resources == nil {
		t.Error("expected non-nil resources")
	}
}

package workspace

import (
	"fmt"
)

// ResourceLimits represents container resource limits
type ResourceLimits struct {
	CPU         string `yaml:"cpu,omitempty"`         // e.g., "1.0", "2"
	Memory      string `yaml:"memory,omitempty"`      // e.g., "512m", "1g"
	MemorySwap  string `yaml:"memory_swap,omitempty"` // e.g., "1g"
	CPUset      string `yaml:"cpuset,omitempty"`      // e.g., "0,1"
	DiskQuota   string `yaml:"disk_quota,omitempty"`  // e.g., "10g"
	PidsLimit   int64  `yaml:"pids_limit,omitempty"`  // e.g., 100
	IOWeight    int    `yaml:"io_weight,omitempty"`   // e.g., 500
	IOReadBPS   int64  `yaml:"io_read_bps,omitempty"`  // bytes per second
	IOWriteBPS  int64  `yaml:"io_write_bps,omitempty"`
	IOReadIOPS  int64  `yaml:"io_read_iops,omitempty"`
	IOWriteIOPS int64  `yaml:"io_write_iops,omitempty"`
}

// ToDockerResources converts to Docker resource config
func (r *ResourceLimits) ToDockerResources() map[string]string {
	res := make(map[string]string)

	if r.CPU != "" {
		res["nano-cpus"] = r.CPU
	}
	if r.Memory != "" {
		res["memory"] = r.Memory
	}
	if r.MemorySwap != "" {
		res["memory-swap"] = r.MemorySwap
	}
	if r.CPUset != "" {
		res["cpuset-cpus"] = r.CPUset
	}
	if r.DiskQuota != "" {
		res["disk-quota"] = r.DiskQuota
	}
	if r.PidsLimit > 0 {
		res["pids-limit"] = fmt.Sprintf("%d", r.PidsLimit)
	}
	if r.IOWeight > 0 {
		res["io-weight"] = fmt.Sprintf("%d", r.IOWeight)
	}
	if r.IOReadBPS > 0 {
		res["io-read-bps"] = fmt.Sprintf("%d", r.IOReadBPS)
	}
	if r.IOWriteBPS > 0 {
		res["io-write-bps"] = fmt.Sprintf("%d", r.IOWriteBPS)
	}
	if r.IOReadIOPS > 0 {
		res["io-read-iops"] = fmt.Sprintf("%d", r.IOReadIOPS)
	}
	if r.IOWriteIOPS > 0 {
		res["io-write-iops"] = fmt.Sprintf("%d", r.IOWriteIOPS)
	}

	return res
}

// Validate validates resource limits
func (r *ResourceLimits) Validate() error {
	// Basic validation - could be extended
	if r.CPU != "" && r.CPU != "0" {
		// Check it's a valid number
		var cpu float64
		if _, err := fmt.Sscanf(r.CPU, "%f", &cpu); err != nil {
			return fmt.Errorf("invalid CPU value: %s", r.CPU)
		}
		if cpu <= 0 || cpu > 1024 {
			return fmt.Errorf("CPU must be between 0 and 1024: %s", r.CPU)
		}
	}

	return nil
}

// DefaultResources returns default resource limits
func DefaultResources() *ResourceLimits {
	return &ResourceLimits{
		CPU:    "1.0",
		Memory: "512m",
	}
}

// GetResourcesFromConfig extracts resource limits from workspace config
func GetResourcesFromConfig(settings map[string]string) *ResourceLimits {
	if settings == nil {
		return DefaultResources()
	}

	limits := &ResourceLimits{
		CPU:    settings["cpu"],
		Memory: settings["memory"],
	}

	if v, ok := settings["cpuset"]; ok {
		limits.CPUset = v
	}
	if v, ok := settings["memory_swap"]; ok {
		limits.MemorySwap = v
	}
	if v, ok := settings["pids_limit"]; ok {
		fmt.Sscanf(v, "%d", &limits.PidsLimit)
	}
	if v, ok := settings["io_weight"]; ok {
		fmt.Sscanf(v, "%d", &limits.IOWeight)
	}

	return limits
}

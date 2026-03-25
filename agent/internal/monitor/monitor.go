package monitor

import (
	"log"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Stats struct {
	CPUUsage    float64
	MemoryUsage float64
}

type Monitor struct{}

func New() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Collect() Stats {
	stats := Stats{}

	// CPU 使用率（采样 200ms）
	percents, err := cpu.Percent(0, false)
	if err == nil && len(percents) > 0 {
		stats.CPUUsage = percents[0]
	} else {
		log.Printf("CPU 采集失败: %v", err)
	}

	// 内存使用率
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		stats.MemoryUsage = vmStat.UsedPercent
	} else {
		log.Printf("内存采集失败: %v", err)
	}

	return stats
}

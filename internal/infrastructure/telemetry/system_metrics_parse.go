package telemetry

import (
	"strconv"
	"strings"
)

// Round-019 machine-wide POSIX CPU/memory parsing (research Decision 5). Kept
// build-tag-free so the parser unit tests (T014) run on any host; the OS
// adapters feed these parsers live readings.

// procCPUTicks is the host CPU tick total + idle tick count parsed from
// /proc/stat's aggregate `cpu ` line.
type procCPUTicks struct {
	total uint64
	idle  uint64
}

// parseProcStatHostCPU parses the aggregate host `cpu ` line of /proc/stat: the
// total is the sum of every tick column and the idle is idle + iowait.
func parseProcStatHostCPU(data string) (procCPUTicks, bool) {
	for _, line := range strings.Split(data, "\n") {
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 { // "cpu" + at least user/nice/system/idle
			return procCPUTicks{}, false
		}
		ticks := make([]uint64, 0, len(fields)-1)
		for _, f := range fields[1:] {
			v, err := strconv.ParseUint(f, 10, 64)
			if err != nil {
				return procCPUTicks{}, false
			}
			ticks = append(ticks, v)
		}
		var total, idle uint64
		for _, v := range ticks {
			total += v
		}
		if len(ticks) > 3 {
			idle += ticks[3] // idle
		}
		if len(ticks) > 4 {
			idle += ticks[4] // iowait
		}
		return procCPUTicks{total: total, idle: idle}, true
	}
	return procCPUTicks{}, false
}

// cpuPercentFromDelta returns the machine-wide CPU percentage over the tick
// delta (Δ of Σcpu − idle). A zero total delta is reported as 0.
func cpuPercentFromDelta(prev, cur procCPUTicks) float64 {
	dTotal := cur.total - prev.total
	if dTotal == 0 {
		return 0
	}
	dIdle := cur.idle - prev.idle
	if dIdle > dTotal {
		dIdle = dTotal // clamp: idle can never exceed the total delta
	}
	return float64(dTotal-dIdle) / float64(dTotal) * 100
}

// parseProcMeminfo parses /proc/meminfo into a used-memory percentage
// ((MemTotal − MemAvailable) / MemTotal · 100).
func parseProcMeminfo(data string) (float64, bool) {
	var total, available uint64
	var haveTotal, haveAvailable bool
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			if v, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				total, haveTotal = v, true
			}
		case "MemAvailable:":
			if v, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				available, haveAvailable = v, true
			}
		}
	}
	if !haveTotal || !haveAvailable || total == 0 {
		return 0, false
	}
	if available > total {
		available = total
	}
	return float64(total-available) / float64(total) * 100, true
}

package hooks

import (
	"strconv"
	"strings"
	"time"
)

const maxPayloadValueLength = 8192

const defaultMaxPayloadBytes = 32768

// PayloadGuard configures payload truncation safeguards.
type PayloadGuard struct {
	MaxValueLength int
	MaxTotalBytes  int
}

// NormalizePayload normalizes hook payload keys to uppercase snake case.
func NormalizePayload(data map[string]string) map[string]string {
	if len(data) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(data))
	for k, v := range data {
		norm := strings.ToUpper(strings.TrimSpace(k))
		norm = strings.ReplaceAll(norm, "-", "_")
		norm = strings.ReplaceAll(norm, " ", "_")
		if norm == "" {
			continue
		}
		out[norm] = trimPayloadValue(v)
	}
	return out
}

// BuildPayload adds standard hook fields and normalizes custom payload keys.
func BuildPayload(event Event, data map[string]string) map[string]string {
	return BuildPayloadGuarded(event, data, PayloadGuard{MaxValueLength: maxPayloadValueLength, MaxTotalBytes: defaultMaxPayloadBytes})
}

// BuildPayloadGuarded adds standard fields and enforces payload-size guards.
func BuildPayloadGuarded(event Event, data map[string]string, guard PayloadGuard) map[string]string {
	if guard.MaxValueLength <= 0 {
		guard.MaxValueLength = maxPayloadValueLength
	}
	if guard.MaxTotalBytes <= 0 {
		guard.MaxTotalBytes = defaultMaxPayloadBytes
	}

	custom := NormalizePayload(data)
	out := make(map[string]string, len(custom)+3)
	out["EVENT"] = string(event)
	out["EVENT_TIME_MS"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	out["SCHEMA"] = "v2"
	total := len(out["EVENT"]) + len(out["EVENT_TIME_MS"]) + len(out["SCHEMA"])
	for k, v := range custom {
		if len(v) > guard.MaxValueLength {
			v = v[:guard.MaxValueLength]
		}
		if total+len(k)+len(v) > guard.MaxTotalBytes {
			out["PAYLOAD_TRUNCATED"] = "1"
			break
		}
		total += len(k) + len(v)
		out[k] = v
	}
	return out
}

func trimPayloadValue(v string) string {
	v = strings.TrimSpace(v)
	if len(v) <= maxPayloadValueLength {
		return v
	}
	return v[:maxPayloadValueLength]
}

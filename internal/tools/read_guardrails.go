package tools

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

const (
	readMetadataStart = "[read_metadata]"
	readMetadataEnd   = "[/read_metadata]"
)

type readMetadata struct {
	Path        string
	MtimeUnixMs int64
	SizeBytes   int64
	HasSize     bool
	Partial     bool
}

type staleReadError struct {
	ReasonCode string
	Message    string
}

func (e staleReadError) Error() string {
	return e.Message + "\n\n" + renderStructuredBlock("stale_read", []structuredField{
		{Key: "ok_to_write", Value: "false"},
		{Key: "reason", Value: e.ReasonCode},
	})
}

func renderReadMetadata(path string, mtime time.Time, sizeBytes int64, isPartial bool, offset, limit int) string {
	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}
	limitField := "all"
	if limit > 0 {
		limitField = strconv.Itoa(limit)
	}
	truncationHint := "none"
	if isPartial {
		truncationHint = "use_offset_and_limit_for_remaining_content"
	}
	return renderStructuredBlock("read_metadata", []structuredField{
		{Key: "path", Value: path},
		{Key: "media_kind", Value: "text"},
		{Key: "mime_type", Value: "text/plain"},
		{Key: "mtime_unix_ms", Value: strconv.FormatInt(mtime.UnixMilli(), 10)},
		{Key: "size_bytes", Value: strconv.FormatInt(sizeBytes, 10)},
		{Key: "offset", Value: strconv.Itoa(offset)},
		{Key: "limit", Value: limitField},
		{Key: "partial", Value: strconv.FormatBool(isPartial)},
		{Key: "truncation_hint", Value: truncationHint},
	})
}

func renderReadAttachmentMetadata(path, mediaKind, mimeType string, mtime time.Time, sizeBytes int64) string {
	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}
	return renderStructuredBlock("read_metadata", []structuredField{
		{Key: "path", Value: path},
		{Key: "media_kind", Value: mediaKind},
		{Key: "mime_type", Value: mimeType},
		{Key: "mtime_unix_ms", Value: strconv.FormatInt(mtime.UnixMilli(), 10)},
		{Key: "size_bytes", Value: strconv.FormatInt(sizeBytes, 10)},
		{Key: "offset", Value: "0"},
		{Key: "limit", Value: "all"},
		{Key: "partial", Value: "false"},
		{Key: "truncation_hint", Value: "none"},
	})
}

func parseReadMetadata(content string) (readMetadata, bool) {
	start := strings.Index(content, readMetadataStart)
	if start < 0 {
		return readMetadata{}, false
	}
	end := strings.Index(content[start:], readMetadataEnd)
	if end < 0 {
		return readMetadata{}, false
	}
	block := content[start+len(readMetadataStart) : start+end]

	meta := readMetadata{}
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "path":
			meta.Path = value
		case "mtime_unix_ms":
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return readMetadata{}, false
			}
			meta.MtimeUnixMs = parsed
		case "size_bytes":
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return readMetadata{}, false
			}
			meta.SizeBytes = parsed
			meta.HasSize = true
		case "partial":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return readMetadata{}, false
			}
			meta.Partial = parsed
		}
	}

	if meta.Path == "" || meta.MtimeUnixMs <= 0 {
		return readMetadata{}, false
	}
	return meta, true
}

func latestReadMetadataForPath(messages []types.Message, path string) (readMetadata, bool) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		for j := len(msg.Content) - 1; j >= 0; j-- {
			meta, ok := parseReadMetadata(msg.Content[j].Content)
			if !ok {
				continue
			}
			if meta.Path == absPath {
				return meta, true
			}
		}
	}

	return readMetadata{}, false
}

func validateReadBeforeModify(messages []types.Message, path string, currentMtime time.Time) error {
	if len(messages) == 0 {
		return nil
	}
	meta, ok := latestReadMetadataForPath(messages, path)
	if !ok {
		return staleReadError{ReasonCode: "not_read", Message: "file has not been read yet. Read it first before writing to it"}
	}
	if meta.Partial {
		return staleReadError{ReasonCode: "partial_read", Message: "file was only partially read. Read the full file before writing to it"}
	}
	if currentMtime.UnixMilli() > meta.MtimeUnixMs {
		return staleReadError{ReasonCode: "mtime_changed", Message: "file has been modified since read. Read it again before writing to it"}
	}
	stat, err := os.Stat(path)
	if err == nil {
		if !meta.HasSize {
			return staleReadError{ReasonCode: "missing_size", Message: "read metadata is stale or incomplete. Read it again before writing to it"}
		}
		if stat.Size() != meta.SizeBytes {
			return staleReadError{ReasonCode: "size_changed", Message: "file size has changed since read. Read it again before writing to it"}
		}
	}
	return nil
}

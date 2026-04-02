package tools

import (
	"strings"
	"testing"
	"time"
)

func TestReadAttachmentMetadataImageAndPDF(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	img, ok := readAttachmentMetadata("/tmp/photo.png", now, 123)
	if !ok {
		t.Fatalf("expected image metadata")
	}
	if !strings.Contains(img, "[read_metadata]") {
		t.Fatalf("expected read_metadata block, got %q", img)
	}
	if !strings.Contains(img, "media_kind: image") {
		t.Fatalf("expected media_kind image, got %q", img)
	}
	if !strings.Contains(img, "mime_type: image/png") {
		t.Fatalf("expected png mime type, got %q", img)
	}
	if !strings.Contains(img, "size_bytes: 123") {
		t.Fatalf("expected size_bytes, got %q", img)
	}

	pdf, ok := readAttachmentMetadata("/tmp/doc.pdf", now, 456)
	if !ok {
		t.Fatalf("expected pdf metadata")
	}
	if !strings.Contains(pdf, "media_kind: pdf") {
		t.Fatalf("expected media_kind pdf, got %q", pdf)
	}
	if !strings.Contains(pdf, "mime_type: application/pdf") {
		t.Fatalf("expected pdf mime type, got %q", pdf)
	}
	if !strings.Contains(pdf, "truncation_hint: none") {
		t.Fatalf("expected truncation_hint none, got %q", pdf)
	}

	if _, ok := readAttachmentMetadata("/tmp/readme.txt", now, 12); ok {
		t.Fatalf("did not expect metadata for txt")
	}
}

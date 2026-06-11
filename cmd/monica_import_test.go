package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/nhymxu/kith-pms/internal/monica"
)

func TestResolveChoiceExplicitModes(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(""))

	yes, err := resolveChoice(reader, "completed", "completed", "skip", "question")
	if err != nil || !yes {
		t.Fatalf("expected completed to resolve true, got %v, %v", yes, err)
	}

	no, err := resolveChoice(reader, "skip", "completed", "skip", "question")
	if err != nil || no {
		t.Fatalf("expected skip to resolve false, got %v, %v", no, err)
	}

	if _, err := resolveChoice(reader, "bad", "completed", "skip", "question"); err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestParseDataURL(t *testing.T) {
	// Valid JPEG data URL (1x1 white JPEG, base64)
	jpegB64 := "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFAABAAAAAAAAAAAAAAAAAAAACf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/xAAUAQEAAAAAAAAAAAAAAAAAAAAA/8QAFBEBAAAAAAAAAAAAAAAAAAAAAP/aAAwDAQACEQMRAD8AJQAB/9k=" //nolint:lll
	dataURL := "data:image/jpeg;base64," + jpegB64

	mimeType, data, err := parseDataURL(dataURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mimeType != "image/jpeg" {
		t.Errorf("expected image/jpeg, got %q", mimeType)
	}

	if len(data) == 0 {
		t.Error("expected non-empty data")
	}

	// Invalid: no data: prefix
	if _, _, err := parseDataURL("notadataurl"); err == nil {
		t.Error("expected error for non-data URL")
	}

	// Invalid: no base64 marker
	if _, _, err := parseDataURL("data:image/jpeg,rawdata"); err == nil {
		t.Error("expected error for missing base64 marker")
	}

	// Invalid base64
	if _, _, err := parseDataURL("data:image/jpeg;base64,!!!invalid!!!"); err == nil {
		t.Error("expected error for invalid base64")
	}

	// Oversized payload (>6.7 MB encoded)
	oversized := "data:image/jpeg;base64," + strings.Repeat("A", 7*1024*1024)
	if _, _, err := parseDataURL(oversized); err == nil {
		t.Error("expected error for oversized data URL")
	}
}

func TestResolveMonicaImportOptionsWithExplicitModes(t *testing.T) {
	export := &monica.Export{
		Contacts: []monica.Contact{
			{Reminders: []monica.MReminder{{Title: "Paused", InitialDate: "2024-06-01", Inactive: true}}},
		},
		AccountJournalEntries: []monica.MAccountJournal{
			{Title: "Private", Content: "Account note", OccurredAtDate: "2024-01-02"},
		},
	}

	options, err := resolveMonicaImportOptions(export, "completed", "unlinked", "skip", "western")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !options.ImportInactiveReminders || !options.ImportAccountJournalEntries {
		t.Fatalf("expected both explicit import options enabled, got %+v", options)
	}

	options, err = resolveMonicaImportOptions(export, "skip", "skip", "skip", "western")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if options.ImportInactiveReminders || options.ImportAccountJournalEntries {
		t.Fatalf("expected both explicit import options disabled, got %+v", options)
	}
}

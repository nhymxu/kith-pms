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

func TestResolveMonicaImportOptionsWithExplicitModes(t *testing.T) {
	export := &monica.Export{
		Contacts:              []monica.Contact{{Reminders: []monica.MReminder{{Title: "Paused", InitialDate: "2024-06-01", Inactive: true}}}},
		AccountJournalEntries: []monica.MAccountJournal{{Title: "Private", Content: "Account note", OccurredAtDate: "2024-01-02"}},
	}

	options, err := resolveMonicaImportOptions(export, "completed", "unlinked")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !options.ImportInactiveReminders || !options.ImportAccountJournalEntries {
		t.Fatalf("expected both explicit import options enabled, got %+v", options)
	}

	options, err = resolveMonicaImportOptions(export, "skip", "skip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if options.ImportInactiveReminders || options.ImportAccountJournalEntries {
		t.Fatalf("expected both explicit import options disabled, got %+v", options)
	}
}

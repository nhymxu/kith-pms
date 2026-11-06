// Package calendar renders birthdays, important dates and reminders as RFC 5545
// iCalendar feeds and manages the secret token that guards the public feed URLs.
package calendar

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	crlf          = "\r\n"
	maxLineOctets = 75
)

// writer accumulates content lines for a VCALENDAR.
type writer struct {
	sb strings.Builder
}

// escapeText escapes a TEXT value per RFC 5545 section 3.3.11.
func escapeText(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		";", `\;`,
		",", `\,`,
		"\r\n", `\n`,
		"\n", `\n`,
		"\r", `\n`,
	)

	return r.Replace(s)
}

// line writes one content line, folded at 75 octets without splitting a rune.
func (w *writer) line(s string) {
	limit := maxLineOctets

	for len(s) > limit {
		cut := limit
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}

		w.sb.WriteString(s[:cut])
		w.sb.WriteString(crlf)
		w.sb.WriteString(" ")

		s = s[cut:]
		// Continuation lines carry a leading space that counts toward the limit.
		limit = maxLineOctets - 1
	}

	w.sb.WriteString(s)
	w.sb.WriteString(crlf)
}

func (w *writer) text(name, value string) {
	if value == "" {
		return
	}

	w.line(name + ":" + escapeText(value))
}

func formatDate(t time.Time) string     { return t.Format("20060102") }
func formatDateTime(t time.Time) string { return t.UTC().Format("20060102T150405Z") }

// event is a single VEVENT. Allday events use VALUE=DATE and an exclusive DTEND.
type event struct {
	uid         string
	summary     string
	description string
	start       time.Time
	allDay      bool
	rrule       string
}

func (w *writer) begin(calName string) {
	w.line("BEGIN:VCALENDAR")
	w.line("VERSION:2.0")
	w.line("PRODID:-//kith-pms//calendar feed//EN")
	w.line("CALSCALE:GREGORIAN")
	w.line("METHOD:PUBLISH")
	w.text("X-WR-CALNAME", calName)
}

func (w *writer) end() { w.line("END:VCALENDAR") }

func (w *writer) event(e event, stamp time.Time) {
	w.line("BEGIN:VEVENT")
	w.line("UID:" + e.uid)
	w.line("DTSTAMP:" + formatDateTime(stamp))

	if e.allDay {
		w.line("DTSTART;VALUE=DATE:" + formatDate(e.start))
		w.line("DTEND;VALUE=DATE:" + formatDate(e.start.AddDate(0, 0, 1)))
	} else {
		w.line("DTSTART:" + formatDateTime(e.start))
	}

	if e.rrule != "" {
		w.line("RRULE:" + e.rrule)
	}

	w.text("SUMMARY", e.summary)
	w.text("DESCRIPTION", e.description)
	w.line("TRANSP:TRANSPARENT")
	w.line("END:VEVENT")
}

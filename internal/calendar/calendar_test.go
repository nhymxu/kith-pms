package calendar

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/settings"
	"github.com/nhymxu/kith-pms/internal/testutil"
)

func TestEscapeText(t *testing.T) {
	got := escapeText("a,b;c\\d\ne")
	want := `a\,b\;c\\d\ne`

	if got != want {
		t.Fatalf("escapeText = %q, want %q", got, want)
	}
}

func TestLineFoldingKeepsRunesIntact(t *testing.T) {
	var w writer

	w.line("SUMMARY:" + strings.Repeat("é", 60))

	for _, l := range strings.Split(strings.TrimSuffix(w.sb.String(), crlf), crlf) {
		if len(l) > maxLineOctets {
			t.Fatalf("line exceeds %d octets: %d", maxLineOctets, len(l))
		}
	}

	unfolded := strings.ReplaceAll(w.sb.String(), crlf+" ", "")
	if unfolded != "SUMMARY:"+strings.Repeat("é", 60)+crlf {
		t.Fatalf("unfolded content differs: %q", unfolded)
	}
}

func seed(t *testing.T) (*Builder, *TokenStore) {
	t.Helper()

	db := testutil.NewDB(t)
	ctx := context.Background()

	stmts := []string{
		`INSERT INTO person (id, name, nickname, date_of_birth) VALUES (1, 'Lan Nguyen', 'Lan', '1990-05-17')`,
		`INSERT INTO person (id, name, date_of_birth) VALUES (2, 'No Year', '--02-29')`,
		`INSERT INTO person (id, name, date_of_birth, archived_at)
		   VALUES (3, 'Archived', '1980-01-01', '2026-01-01T00:00:00Z')`,
		`INSERT INTO person (id, name) VALUES (4, 'No Dob')`,
		`INSERT INTO important_date (person_id, kind, label, date_value, recurring, notes, position)
		   VALUES (1, 'anniversary', 'Wedding', '2015-06-01', 1, 'with, comma', 0)`,
		`INSERT INTO reminder (title, notes, due_date, person_id, completed, recurrence_rule)
		   VALUES ('Call Lan', 'about trip', '2026-12-01T00:00:00Z', 1, 0, '{"type":"weekly"}')`,
		`INSERT INTO important_date (person_id, kind, label, date_value, recurring, notes, position)
		   VALUES (1, '', '', '--05-01', 0, '', 1)`,
		`INSERT INTO person (id, name, deleted_at) VALUES (5, 'Gone', '2026-01-01T00:00:00Z')`,
		`INSERT INTO reminder (title, due_date, person_id, completed) VALUES ('Gone rem', '2026-12-03T00:00:00Z', 5, 0)`,
		`INSERT INTO reminder (title, due_date, completed) VALUES ('Done thing', '2026-12-02T00:00:00Z', 1)`,
		`INSERT INTO reminder (title, due_date, completed, recurrence_rule)
		   VALUES ('Bday rem', '2026-05-10T00:00:00Z', 0, '{"type":"birthday"}')`,
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			t.Fatalf("seed %q: %v", s, err)
		}
	}

	return &Builder{
			DB:        db,
			Reminders: reminders.NewRepo(db),
			Now:       func() time.Time { return time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC) },
		},
		&TokenStore{Repo: settings.NewRepo(db)}
}

func TestBirthdaysFeed(t *testing.T) {
	b, _ := seed(t)

	out, err := b.Build(context.Background(), "t", PartBirthdays, PartDates)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"BEGIN:VCALENDAR\r\n",
		"UID:kith-birthday-1@kith\r\n",
		"DTSTART;VALUE=DATE:19900517\r\n",
		"RRULE:FREQ=YEARLY\r\n",
		"SUMMARY:Lan Nguyen (Lan)'s birthday\r\n",
		"DESCRIPTION:Born 1990\r\n",
		"UID:kith-birthday-2@kith\r\n",
		"DTSTART;VALUE=DATE:20000229\r\n",
		"SUMMARY:Wedding: Lan Nguyen (Lan)\r\n",
		`DESCRIPTION:Since 2015\nwith\, comma` + "\r\n",
		"END:VCALENDAR\r\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}

	if !strings.Contains(out, "SUMMARY:Date: Lan Nguyen (Lan)\r\n") {
		t.Errorf("empty-kind date not rendered with fallback label:\n%s", out)
	}

	for _, bad := range []string{"Archived", "No Dob", "Call Lan"} {
		if strings.Contains(out, bad) {
			t.Errorf("unexpected %q in birthdays feed", bad)
		}
	}
}

func TestRemindersFeedSkipsCompletedAndBirthday(t *testing.T) {
	b, _ := seed(t)

	out, err := b.Build(context.Background(), "t", PartReminders)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "SUMMARY:Call Lan\r\n") || !strings.Contains(out, "RRULE:FREQ=WEEKLY\r\n") {
		t.Errorf("pending weekly reminder missing:\n%s", out)
	}

	for _, bad := range []string{"Done thing", "Bday rem", "birthday", "Gone rem"} {
		if strings.Contains(out, bad) {
			t.Errorf("unexpected %q in reminders feed", bad)
		}
	}
}

func TestCombinedFeedHasBoth(t *testing.T) {
	b, _ := seed(t)

	out, err := b.Build(context.Background(), "t", AllParts...)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "kith-birthday-1@kith") || !strings.Contains(out, "kith-reminder-1@kith") {
		t.Errorf("combined feed incomplete:\n%s", out)
	}
}

func TestReminderRRule(t *testing.T) {
	sun := 0
	end := time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC)

	rule := func(rr reminders.RecurrenceRule, endDate *time.Time) reminders.Reminder {
		return reminders.Reminder{RecurrenceRule: &rr, RecurrenceEndDate: endDate}
	}

	cases := []struct {
		name   string
		r      reminders.Reminder
		allDay bool
		want   string
	}{
		{"none", reminders.Reminder{}, true, ""},
		{"custom weeks", rule(reminders.RecurrenceRule{Type: "custom", Interval: 2, Unit: "weeks"}, nil), true,
			"FREQ=WEEKLY;INTERVAL=2"},
		{"day of week", rule(reminders.RecurrenceRule{Type: "day_of_week", DayOfWeek: &sun}, nil), true,
			"FREQ=WEEKLY;BYDAY=SU"},
		{"relative contact", rule(reminders.RecurrenceRule{Type: "relative_contact", Interval: 3}, nil), true, ""},
		{"until date", rule(reminders.RecurrenceRule{Type: "daily"}, &end), true, "FREQ=DAILY;UNTIL=20270105"},
		{"until timed", rule(reminders.RecurrenceRule{Type: "daily"}, &end), false,
			"FREQ=DAILY;UNTIL=20270105T235959Z"},
	}

	for _, c := range cases {
		if got := reminderRRule(c.r, c.allDay); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestTokenLifecycle(t *testing.T) {
	_, ts := seed(t)
	ctx := context.Background()

	if ok, _ := ts.Valid(ctx, "anything"); ok {
		t.Fatal("token valid before creation")
	}

	tok, err := ts.Get(ctx)
	if err != nil || len(tok) < 40 {
		t.Fatalf("Get = %q, %v", tok, err)
	}

	again, _ := ts.Get(ctx)
	if again != tok {
		t.Fatal("Get is not stable")
	}

	if ok, _ := ts.Valid(ctx, tok); !ok {
		t.Fatal("current token rejected")
	}

	if ok, _ := ts.Valid(ctx, ""); ok {
		t.Fatal("empty token accepted")
	}

	next, err := ts.Regenerate(ctx)
	if err != nil || next == tok {
		t.Fatalf("Regenerate = %q, %v", next, err)
	}

	if ok, _ := ts.Valid(ctx, tok); ok {
		t.Fatal("old token still valid after regenerate")
	}
}

func TestPartsSelectContent(t *testing.T) {
	b, _ := seed(t)

	out, err := b.Build(context.Background(), "t", PartBirthdays)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "kith-birthday-1@kith") || strings.Contains(out, "kith-date-") ||
		strings.Contains(out, "kith-reminder-") {
		t.Errorf("birthdays-only feed has wrong content:\n%s", out)
	}

	if got := ParseParts("reminders, bogus,reminders,dates"); len(got) != 2 {
		t.Errorf("ParseParts = %v, want [reminders dates]", got)
	}
}

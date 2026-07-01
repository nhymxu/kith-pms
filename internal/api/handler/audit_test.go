package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/settings"
)

func TestAuditCleanup_RetentionDisabled(t *testing.T) {
	db := openTestDB(t)
	auditSvc := audit.NewService(db)
	settingsSvc := settings.NewService(db)

	h := &handler.AuditAPI{Svc: auditSvc, SettingsSvc: settingsSvc}
	e := newTestEcho()

	req := jsonRequest(http.MethodPost, "/v1/audit/cleanup", "")
	rec := execHandler(e, req, nil, h.Cleanup)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body["deleted"] != 0 {
		t.Errorf("want deleted=0, got %d", body["deleted"])
	}
}

func TestAuditCleanup_DeletesEntries(t *testing.T) {
	db := openTestDB(t)
	auditSvc := audit.NewService(db)
	settingsSvc := settings.NewService(db)

	// seed one old entry (91 days ago) and one recent entry
	old := time.Now().AddDate(0, 0, -91).UTC().Format("2006-01-02T15:04:05Z")

	recent := time.Now().AddDate(0, 0, -1).UTC().Format("2006-01-02T15:04:05Z")
	for _, ts := range []string{old, recent} {
		if _, err := db.ExecContext(
			t.Context(),
			`INSERT INTO audit_log (entity_type, entity_id, entity_name, action, created_at) VALUES ('person', 1, 'Test', 'create', ?)`, //nolint:lll
			ts,
		); err != nil {
			t.Fatalf("insert audit: %v", err)
		}
	}

	// set retention to 30 days
	if _, err := settingsSvc.Update(t.Context(), settings.UserSettings{
		DateFormat:            "YYYY-MM-DD",
		TimeFormat:            "24h",
		Timezone:              "UTC",
		AuditLogRetentionDays: 30,
		NetworkColorBy:        "labels",
		NetworkOnlyMineDepth:  "direct",
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	h := &handler.AuditAPI{Svc: auditSvc, SettingsSvc: settingsSvc}
	e := newTestEcho()

	req := jsonRequest(http.MethodPost, "/v1/audit/cleanup", "")
	rec := execHandler(e, req, nil, h.Cleanup)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body["deleted"] != 1 {
		t.Errorf("want deleted=1, got %d", body["deleted"])
	}
}

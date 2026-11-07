package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/internal/settings"
)

func TestSettingsPutWithoutNotifyHourKeepsStoredValue(t *testing.T) {
	svc := settings.NewService(openTestDB(t))
	ctx := context.Background()

	cur, _ := svc.Get(ctx)
	cur.NotifyHour = 18

	if _, err := svc.Update(ctx, cur); err != nil {
		t.Fatal(err)
	}

	// a client that predates notify_hour sends everything else
	raw, _ := json.Marshal(cur)

	var body map[string]any

	_ = json.Unmarshal(raw, &body)
	delete(body, "notify_hour")

	b, _ := json.Marshal(body)
	h := &handler.SettingsAPI{Svc: svc}

	rec := execHandler(newTestEcho(), jsonRequest(http.MethodPut, "/v1/settings", string(b)), nil, h.Update)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT: %d %s", rec.Code, rec.Body.String())
	}

	if got, _ := svc.Get(ctx); got.NotifyHour != 18 {
		t.Fatalf("notify_hour reset to %d", got.NotifyHour)
	}
}

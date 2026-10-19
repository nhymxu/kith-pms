package handler_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/settings"
)

type errBody struct {
	Error  string         `json:"error"`
	Code   string         `json:"code"`
	Params map[string]any `json:"params"`
}

func decodeErrBody(t *testing.T, raw []byte) errBody {
	t.Helper()

	var b errBody
	if err := json.Unmarshal(raw, &b); err != nil {
		t.Fatalf("decode: %v", err)
	}

	return b
}

func TestSettingsUpdate_InvalidLanguage_ReturnsCode(t *testing.T) {
	h := &handler.SettingsAPI{Svc: settings.NewService(openTestDB(t))}
	e := newTestEcho()

	body := `{"date_format":"YYYY-MM-DD","time_format":"24h","timezone":"UTC","network_color_by":"labels",` +
		`"network_only_mine_depth":"direct","default_people_sort":"name","default_page_size":25,` +
		`"dashboard_favorites_count":5,"dashboard_last_contact_count":5,"theme":"quiet-ink","nav_layout":"top",` +
		`"number_format":"1,234.56","relationship_display":"plain","language":"xx","search_scope":["people"]}`

	rec := execHandler(e, jsonRequest(http.MethodPut, "/v1/settings", body), nil, h.Update)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status: want 422, got %d", rec.Code)
	}

	got := decodeErrBody(t, rec.Body.Bytes())
	if got.Code != "settings_language_invalid" || got.Error == "" {
		t.Errorf("want code settings_language_invalid with english text, got %+v", got)
	}
}

func TestSettingsUpdate_BadBody_ReturnsInvalidRequestBodyCode(t *testing.T) {
	h := &handler.SettingsAPI{Svc: settings.NewService(openTestDB(t))}
	e := newTestEcho()

	rec := execHandler(e, jsonRequest(http.MethodPut, "/v1/settings", "{"), nil, h.Update)

	got := decodeErrBody(t, rec.Body.Bytes())
	if rec.Code != http.StatusBadRequest || got.Code != "invalid_request_body" {
		t.Errorf("want 400 invalid_request_body, got %d %+v", rec.Code, got)
	}
}

func TestPeopleLabelCreate_NameTooLong_ReturnsCodeAndMax(t *testing.T) {
	h := &handler.PeopleLabelsCRUD{Svc: people.NewLabelService(openTestDB(t))}
	e := newTestEcho()

	name := strings.Repeat("a", people.MaxLabelNameLen+1)
	body := `{"name":"` + name + `","color":"#a1b2c3"}`

	rec := execHandler(e, jsonRequest(http.MethodPost, "/v1/people/labels", body), nil, h.Create)

	got := decodeErrBody(t, rec.Body.Bytes())
	if rec.Code != http.StatusUnprocessableEntity || got.Code != "label_name_too_long" {
		t.Fatalf("want 422 label_name_too_long, got %d %+v", rec.Code, got)
	}

	if got.Params["max"] != float64(people.MaxLabelNameLen) {
		t.Errorf("params.max: want %d, got %v", people.MaxLabelNameLen, got.Params["max"])
	}
}

package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nhymxu/kith-pms/internal/people"
)

func TestPeopleFormDynamicRowsUseCurrentRowCounts(t *testing.T) {
	var buf bytes.Buffer
	err := PeopleForm(PeopleFormParams{
		Person: people.Person{ID: 42, Name: "Ada"},
		Contacts: []people.ContactInfo{
			{Type: "email", Value: "ada@example.com"},
		},
		IsEdit: true,
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render PeopleForm: %v", err)
	}

	html := buf.String()
	want := []string{
		`hx-vals="js:{count: document.querySelectorAll('#contacts-table tbody tr').length}"`,
		`hx-vals="js:{count: document.querySelectorAll('#locations-table tbody tr').length}"`,
		`hx-vals="js:{index: document.querySelectorAll('#dates-table tbody tr').length}"`,
		`<th class="pb-1 font-medium w-20">Action</th>`,
		`onclick="this.closest('tr').remove()"`,
	}
	for _, marker := range want {
		if !strings.Contains(html, marker) {
			t.Errorf("rendered form missing %s", marker)
		}
	}
}

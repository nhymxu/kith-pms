package people_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/uptrace/bun"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/people"
)

// openTestDB opens an in-memory SQLite database and runs all migrations.
func openTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db, err := internaldb.Open(":memory:", 1)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func newSvc(t *testing.T) *people.Service {
	t.Helper()
	return people.NewService(openTestDB(t))
}

// ---- helpers ----------------------------------------------------------------

func mustCreate(
	t *testing.T,
	svc *people.Service,
	name string,
	contacts []people.ContactInfo,
	locations []people.Location,
) int64 {
	t.Helper()

	id, err := svc.Create(context.Background(), people.Person{Name: name}, contacts, locations)
	if err != nil {
		t.Fatalf("Create(%q): %v", name, err)
	}

	return id
}

// ---- tests ------------------------------------------------------------------

func TestCreate_GetRoundtrip(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	dobTime := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	dob := people.NewDateOnly(dobTime)
	p := people.Person{
		Prefix:      "Dr",
		Name:        "Alice Example",
		Nickname:    "Ali",
		DateOfBirth: &dob,
		OtherNotes:  "met at conference",
	}
	contacts := []people.ContactInfo{
		{Type: "email", Value: "alice@example.com", Label: "work"},
		{Type: "phone", Value: "+1-555-0101", Label: "mobile"},
	}
	locations := []people.Location{
		{Type: "home", City: "Berlin", Country: "DE"},
		{Type: "work", Address: "123 Main St", City: "Hamburg", Country: "DE", PostalCode: "20095"},
	}

	id, err := svc.Create(ctx, p, contacts, locations)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got == nil {
		t.Fatal("Get returned nil")
		return
	}

	// Core fields.
	if got.Name != p.Name {
		t.Errorf("Name: got %q, want %q", got.Name, p.Name)
	}

	if got.Prefix != p.Prefix {
		t.Errorf("Prefix: got %q, want %q", got.Prefix, p.Prefix)
	}

	if got.Nickname != p.Nickname {
		t.Errorf("Nickname: got %q, want %q", got.Nickname, p.Nickname)
	}

	if got.DateOfBirth == nil {
		t.Error("DateOfBirth: got nil, want non-nil")
	} else if got.DateOfBirth.String() != dob.String() {
		t.Errorf("DateOfBirth: got %v, want %v", got.DateOfBirth, dob)
	}

	// Contacts.
	if len(got.Contacts) != 2 {
		t.Fatalf("Contacts: got %d, want 2", len(got.Contacts))
	}

	if got.Contacts[0].Value != "alice@example.com" {
		t.Errorf("Contacts[0].Value: got %q", got.Contacts[0].Value)
	}

	if got.Contacts[1].Type != "phone" {
		t.Errorf("Contacts[1].Type: got %q", got.Contacts[1].Type)
	}

	// Locations.
	if len(got.Locations) != 2 {
		t.Fatalf("Locations: got %d, want 2", len(got.Locations))
	}

	if got.Locations[0].City != "Berlin" {
		t.Errorf("Locations[0].City: got %q", got.Locations[0].City)
	}

	if got.Locations[1].PostalCode != "20095" {
		t.Errorf("Locations[1].PostalCode: got %q", got.Locations[1].PostalCode)
	}
}

func TestUpdate_ReplaceAll(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	// Create with 2 contacts.
	id, err := svc.Create(ctx,
		people.Person{Name: "Bob"},
		[]people.ContactInfo{
			{Type: "email", Value: "bob@old.com"},
			{Type: "phone", Value: "+1-555-0000"},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Update with only 1 different contact — replace-all should remove the other.
	err = svc.Update(ctx,
		people.Person{ID: id, Name: "Bob Updated"},
		[]people.ContactInfo{
			{Type: "email", Value: "bob@new.com", Label: "personal"},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}

	if got.Name != "Bob Updated" {
		t.Errorf("Name: got %q, want %q", got.Name, "Bob Updated")
	}

	if len(got.Contacts) != 1 {
		t.Fatalf("Contacts after replace-all: got %d, want 1", len(got.Contacts))
	}

	if got.Contacts[0].Value != "bob@new.com" {
		t.Errorf("Contacts[0].Value: got %q, want %q", got.Contacts[0].Value, "bob@new.com")
	}
}

func TestDelete_Cascade(t *testing.T) {
	db := openTestDB(t)
	// Enable foreign keys on this connection explicitly.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("pragma: %v", err)
	}

	svc := people.NewService(db)
	ctx := context.Background()

	id, err := svc.Create(ctx,
		people.Person{Name: "Carol"},
		[]people.ContactInfo{{Type: "email", Value: "carol@test.com"}},
		[]people.Location{{Type: "home", City: "Paris"}},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Person should be gone.
	got, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}

	if got != nil {
		t.Error("expected nil person after delete, got non-nil")
	}

	// Contacts and locations should cascade-delete.
	var contactCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM contact_info WHERE person_id = ?`, id).
		Scan(&contactCount); err != nil {
		t.Fatalf("count contacts: %v", err)
	}

	if contactCount != 0 {
		t.Errorf("contact_info not cascaded: got %d rows", contactCount)
	}

	var locationCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM location WHERE person_id = ?`, id).
		Scan(&locationCount); err != nil {
		t.Fatalf("count locations: %v", err)
	}

	if locationCount != 0 {
		t.Errorf("location not cascaded: got %d rows", locationCount)
	}
}

func TestList_Search(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	mustCreate(t, svc, "Alice Wonderland", nil, nil)
	mustCreate(t, svc, "Bob Builder", nil, nil)
	mustCreate(t, svc, "alice cooper", nil, nil) // tests case-insensitive search

	// Search "alice" should match two people (case-insensitive via name_lower).
	results, err := svc.List(ctx, people.ListParams{Query: "alice", PageSize: 50})
	if err != nil {
		t.Fatalf("List search: %v", err)
	}

	if len(results.Items) != 2 {
		t.Errorf("List search 'alice': got %d results, want 2", len(results.Items))
	}

	// Empty query should return all.
	all, err := svc.List(ctx, people.ListParams{PageSize: 50})
	if err != nil {
		t.Fatalf("List all: %v", err)
	}

	if len(all.Items) != 3 {
		t.Errorf("List all: got %d results, want 3", len(all.Items))
	}
}

func TestList_Pagination(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		mustCreate(t, svc, "Person "+string(rune('A'+i)), nil, nil)
	}

	page1, err := svc.List(ctx, people.ListParams{Page: 1, PageSize: 3})
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}

	if len(page1.Items) != 3 {
		t.Errorf("page 1: got %d, want 3", len(page1.Items))
	}

	page2, err := svc.List(ctx, people.ListParams{Page: 2, PageSize: 3})
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}

	if len(page2.Items) != 2 {
		t.Errorf("page 2: got %d, want 2", len(page2.Items))
	}
}

func TestGetSelf_NoneSet(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	got, err := svc.GetSelf(ctx)
	if err != nil {
		t.Fatalf("GetSelf: %v", err)
	}

	if got != nil {
		t.Fatalf("GetSelf: got %#v, want nil", got)
	}
}

func TestSetSelf_AndGet(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	aliceID := mustCreate(t, svc, "Alice", nil, nil)
	bobID := mustCreate(t, svc, "Bob", nil, nil)

	if err := svc.SetSelf(ctx, aliceID); err != nil {
		t.Fatalf("SetSelf alice: %v", err)
	}

	got, err := svc.GetSelf(ctx)
	if err != nil {
		t.Fatalf("GetSelf alice: %v", err)
	}

	if got == nil || got.ID != aliceID {
		t.Fatalf("GetSelf alice: got %#v, want id %d", got, aliceID)
	}

	if err := svc.SetSelf(ctx, bobID); err != nil {
		t.Fatalf("SetSelf bob: %v", err)
	}

	got, err = svc.GetSelf(ctx)
	if err != nil {
		t.Fatalf("GetSelf bob: %v", err)
	}

	if got == nil || got.ID != bobID {
		t.Fatalf("GetSelf bob: got %#v, want id %d", got, bobID)
	}
}

func TestSetSelf_UnknownID(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	if err := svc.SetSelf(ctx, 9999); err == nil {
		t.Fatal("SetSelf unknown ID: got nil error, want error")
	}
}

func TestSetFavorite_Toggle(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	id := mustCreate(t, svc, "Favorite Target", nil, nil)

	if err := svc.SetFavorite(ctx, id, true); err != nil {
		t.Fatalf("SetFavorite true: %v", err)
	}

	got, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if !got.IsFavorite {
		t.Fatalf("IsFavorite: got false, want true")
	}

	if err := svc.SetFavorite(ctx, id, false); err != nil {
		t.Fatalf("SetFavorite false: %v", err)
	}

	got, err = svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.IsFavorite {
		t.Fatalf("IsFavorite: got true, want false")
	}
}

func TestSetFavorite_NotFound(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	err := svc.SetFavorite(ctx, 9999, true)
	if err == nil {
		t.Fatal("SetFavorite unknown ID: got nil error, want error")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("SetFavorite unknown ID: got %q, want error containing 'not found'", err)
	}
}

func TestList_FavoriteOnly(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	aliceID := mustCreate(t, svc, "Alice", nil, nil)
	mustCreate(t, svc, "Bob", nil, nil)
	mustCreate(t, svc, "Carol", nil, nil)

	if err := svc.SetFavorite(ctx, aliceID, true); err != nil {
		t.Fatalf("SetFavorite: %v", err)
	}

	results, err := svc.List(ctx, people.ListParams{FavoriteOnly: true, PageSize: 50})
	if err != nil {
		t.Fatalf("List favorite_only: %v", err)
	}

	if len(results.Items) != 1 {
		t.Fatalf("List favorite_only: got %d results, want 1", len(results.Items))
	}

	if results.Items[0].ID != aliceID {
		t.Fatalf("List favorite_only: got id %d, want %d", results.Items[0].ID, aliceID)
	}
}

func TestList_FavoriteFirst_CombinesWithPrimarySort(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	// Create two people: Bob (unfavorited), Zoe (favorited).
	// Deliberately reverse-alphabetical to prove ordering isn't accidentally name-ascending.
	bobID := mustCreate(t, svc, "Bob", nil, nil)
	zoeID := mustCreate(t, svc, "Zoe", nil, nil)

	if err := svc.SetFavorite(ctx, zoeID, true); err != nil {
		t.Fatalf("SetFavorite Zoe: %v", err)
	}

	// Test 1: FavoriteFirst with "name" sort.
	// Zoe is favorite, Bob is not. Zoe should be first despite B < Z alphabetically.
	results, err := svc.List(ctx, people.ListParams{FavoriteFirst: true, Sort: "name", PageSize: 50})
	if err != nil {
		t.Fatalf("List FavoriteFirst=true, Sort=name: %v", err)
	}

	if len(results.Items) != 2 {
		t.Fatalf("List: got %d results, want 2", len(results.Items))
	}

	if results.Items[0].Name != "Zoe" {
		t.Fatalf("List FavoriteFirst=true, Sort=name: items[0]=%q, want Zoe", results.Items[0].Name)
	}

	// Test 2: FavoriteFirst with "-last_contact" sort.
	// Set distinct LastContactAt so Bob would sort first under plain -last_contact.
	bobTime := time.Now().UTC()

	zoeTime := bobTime.Add(-24 * time.Hour) // Zoe is 1 day older
	if err := svc.People.UpdateLastContact(ctx, svc.DB, bobID, bobTime); err != nil {
		t.Fatalf("UpdateLastContact Bob: %v", err)
	}

	if err := svc.People.UpdateLastContact(ctx, svc.DB, zoeID, zoeTime); err != nil {
		t.Fatalf("UpdateLastContact Zoe: %v", err)
	}

	results, err = svc.List(ctx, people.ListParams{FavoriteFirst: true, Sort: "-last_contact", PageSize: 50})
	if err != nil {
		t.Fatalf("List FavoriteFirst=true, Sort=-last_contact: %v", err)
	}

	if len(results.Items) != 2 {
		t.Fatalf("List: got %d results, want 2", len(results.Items))
	}

	// Zoe is still first because FavoriteFirst takes precedence.
	if results.Items[0].Name != "Zoe" {
		t.Fatalf("List FavoriteFirst=true, Sort=-last_contact: items[0]=%q, want Zoe", results.Items[0].Name)
	}

	// Test 3: FavoriteFirst=false with "name" sort.
	// Bob should be first now (plain alphabetical).
	results, err = svc.List(ctx, people.ListParams{FavoriteFirst: false, Sort: "name", PageSize: 50})
	if err != nil {
		t.Fatalf("List FavoriteFirst=false, Sort=name: %v", err)
	}

	if len(results.Items) != 2 {
		t.Fatalf("List: got %d results, want 2", len(results.Items))
	}

	if results.Items[0].Name != "Bob" {
		t.Fatalf("List FavoriteFirst=false, Sort=name: items[0]=%q, want Bob", results.Items[0].Name)
	}
}

func TestList_LegacyFavoriteSortValue_FallsBackGracefully(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	// Create a few people.
	mustCreate(t, svc, "Carol", nil, nil)
	mustCreate(t, svc, "Alice", nil, nil)
	mustCreate(t, svc, "Bob", nil, nil)

	// Call List with the old "-favorite" sort value (now falls through to default case).
	// Should not error and should return results ordered by name_lower ASC (the default).
	results, err := svc.List(ctx, people.ListParams{Sort: "-favorite", PageSize: 50})
	if err != nil {
		t.Fatalf("List sort -favorite: %v", err)
	}

	if len(results.Items) != 3 {
		t.Fatalf("List: got %d results, want 3", len(results.Items))
	}

	// Should be ordered by name_lower ASC (default), not by favorite status.
	// Expected order: Alice, Bob, Carol
	expectedOrder := []string{"Alice", "Bob", "Carol"}
	for i, expected := range expectedOrder {
		if results.Items[i].Name != expected {
			t.Fatalf("List sort -favorite: items[%d]=%q, want %q", i, results.Items[i].Name, expected)
		}
	}
}

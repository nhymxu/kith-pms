package gifts

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if _, err := db.Exec(`CREATE TABLE person (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`); err != nil {
		t.Fatalf("create person table: %v", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE gift (
			id              INTEGER PRIMARY KEY,
			person_id       INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
			title           TEXT    NOT NULL,
			direction       TEXT    NOT NULL DEFAULT 'planned',
			date            TEXT,
			notes           TEXT    NOT NULL DEFAULT '',
			amount_cents    INTEGER,
			currency        TEXT    NOT NULL DEFAULT 'USD',
			debt_type       TEXT    NOT NULL DEFAULT '',
			image_path      TEXT    NOT NULL DEFAULT '',
			image_mime_type TEXT    NOT NULL DEFAULT '',
			created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);
		CREATE INDEX idx_gift_person    ON gift(person_id);
		CREATE INDEX idx_gift_direction ON gift(direction);
	`); err != nil {
		t.Fatalf("create gift table: %v", err)
	}

	return db
}

func insertPerson(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()

	res, err := db.ExecContext(context.Background(), "INSERT INTO person (name) VALUES (?)", name)
	if err != nil {
		t.Fatalf("insert person: %v", err)
	}

	id, _ := res.LastInsertId()

	return id
}

func TestGiftsCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Alice")

	cents := int64(2500)
	g := &Gift{
		PersonID:    personID,
		Title:       "Birthday book",
		Direction:   DirectionGiven,
		Date:        "2026-04-15",
		Notes:       "A great novel",
		AmountCents: &cents,
		Currency:    "USD",
		DebtType:    DebtNone,
	}

	id, err := svc.Create(ctx, g)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if id <= 0 {
		t.Fatalf("Create returned id=%d, want >0", id)
	}

	got, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Title != "Birthday book" {
		t.Errorf("Title = %q, want Birthday book", got.Title)
	}

	if got.Direction != DirectionGiven {
		t.Errorf("Direction = %q, want given", got.Direction)
	}

	if got.Date != "2026-04-15" {
		t.Errorf("Date = %q, want 2026-04-15", got.Date)
	}

	if got.AmountCents == nil || *got.AmountCents != 2500 {
		t.Errorf("AmountCents = %v, want 2500", got.AmountCents)
	}

	got.Title = "Updated book"
	got.Direction = DirectionReceived

	got.AmountCents = nil
	if err := svc.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if updated.Title != "Updated book" {
		t.Errorf("updated Title = %q, want Updated book", updated.Title)
	}

	if updated.AmountCents != nil {
		t.Errorf("updated AmountCents = %v, want nil", updated.AmountCents)
	}

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := svc.GetByID(ctx, id); err != sql.ErrNoRows {
		t.Errorf("GetByID after delete: got %v, want sql.ErrNoRows", err)
	}
}

func TestGiftsListFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Bob")

	cents := int64(1000)
	directions := []Direction{DirectionGiven, DirectionReceived, DirectionPlanned}

	debtTypes := []DebtType{DebtIOwe, DebtNone, DebtNone}
	for i, d := range directions {
		_, err := svc.Create(ctx, &Gift{
			PersonID:    personID,
			Title:       string(d) + " gift",
			Direction:   d,
			DebtType:    debtTypes[i],
			AmountCents: &cents,
			Currency:    "USD",
		})
		if err != nil {
			t.Fatalf("Create %s: %v", d, err)
		}
	}

	all, err := svc.List(ctx, ListParams{PageSize: 50, Page: 1})
	if err != nil {
		t.Fatalf("List all: %v", err)
	}

	if len(all.Items) != 3 {
		t.Errorf("List all: got %d, want 3", len(all.Items))
	}

	given, err := svc.List(ctx, ListParams{Direction: DirectionGiven, PageSize: 50, Page: 1})
	if err != nil {
		t.Fatalf("List given: %v", err)
	}

	if len(given.Items) != 1 {
		t.Errorf("List direction=given: got %d, want 1", len(given.Items))
	}

	owe, err := svc.List(ctx, ListParams{DebtType: DebtIOwe, PageSize: 50, Page: 1})
	if err != nil {
		t.Fatalf("List i_owe: %v", err)
	}

	if len(owe.Items) != 1 {
		t.Errorf("List debt_type=i_owe: got %d, want 1", len(owe.Items))
	}
}

func TestGiftsPersonCascadeDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Carol")

	_, err := svc.Create(ctx, &Gift{
		PersonID:  personID,
		Title:     "Flowers",
		Direction: DirectionGiven,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := db.ExecContext(ctx, "DELETE FROM person WHERE id = ?", personID); err != nil {
		t.Fatalf("delete person: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM gift WHERE person_id = ?", personID).
		Scan(&count); err != nil {
		t.Fatalf("count gifts: %v", err)
	}

	if count != 0 {
		t.Errorf("got %d gifts after person delete, want 0", count)
	}
}

func TestGiftsDebtTracking(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Dave")

	cents := int64(5000)

	id, err := svc.Create(ctx, &Gift{
		PersonID:    personID,
		Title:       "Loan",
		Direction:   DirectionGiven,
		AmountCents: &cents,
		Currency:    "USD",
		DebtType:    DebtIOwe,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	g, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if !g.IsMoney() {
		t.Error("IsMoney() = false, want true")
	}

	if g.DisplayAmount() != "USD 50" {
		t.Errorf("DisplayAmount() = %q, want USD 50", g.DisplayAmount())
	}

	g.DebtType = DebtNone

	g.AmountCents = nil
	if err := svc.Update(ctx, g); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, _ := svc.GetByID(ctx, id)
	if updated.IsMoney() {
		t.Error("IsMoney() after clearing debt = true, want false")
	}
}

func TestGiftsImageFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Eve")

	id, err := svc.Create(ctx, &Gift{
		PersonID:  personID,
		Title:     "Camera",
		Direction: DirectionReceived,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Manually set image fields via repo to simulate an upload.
	tx, _ := db.BeginTx(ctx, nil)

	repo := NewRepo(db)
	if err := repo.UpdateImage(ctx, tx, id, "gifts/1/abc.jpg", "image/jpeg"); err != nil {
		tx.Rollback()
		t.Fatalf("UpdateImage: %v", err)
	}

	tx.Commit()

	g, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if !g.HasImage() {
		t.Error("HasImage() = false, want true")
	}

	if g.ImagePath != "gifts/1/abc.jpg" {
		t.Errorf("ImagePath = %q, want gifts/1/abc.jpg", g.ImagePath)
	}

	// Clear image via repo.
	tx2, _ := db.BeginTx(ctx, nil)
	if err := repo.UpdateImage(ctx, tx2, id, "", ""); err != nil {
		tx2.Rollback()
		t.Fatalf("UpdateImage clear: %v", err)
	}

	tx2.Commit()

	g2, _ := svc.GetByID(ctx, id)
	if g2.HasImage() {
		t.Error("HasImage() after clearing = true, want false")
	}
}

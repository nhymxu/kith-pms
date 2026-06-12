package gifts

import (
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type Direction string

const (
	DirectionGiven    Direction = "given"
	DirectionReceived Direction = "received"
	DirectionPlanned  Direction = "planned"
)

type DebtType string

const (
	DebtNone    DebtType = ""
	DebtIOwe    DebtType = "i_owe"
	DebtTheyOwe DebtType = "they_owe"
)

type Gift struct {
	bun.BaseModel `bun:"table:gift,alias:g"`

	ID          int64     `bun:",pk,autoincrement" json:"id"`
	PersonID    int64     `bun:"person_id"         json:"person_id"`
	Title       string    `bun:"title"             json:"title"`
	Direction   Direction `bun:"direction"         json:"direction"`
	Date        string    `bun:"date,nullzero"     json:"date"` // YYYY-MM-DD or ""; nullzero stores "" as NULL
	Notes       string    `bun:"notes"             json:"notes"`
	AmountCents *int64    `bun:"amount_cents"      json:"amount_cents"`
	Currency    string    `bun:"currency"          json:"currency"`
	DebtType    DebtType  `bun:"debt_type"         json:"debt_type"`
	ImagePath   string    `bun:"image_path"        json:"image_path"`
	CreatedAt   time.Time `bun:"created_at"        json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"        json:"updated_at"`
}

type GiftWithPerson struct {
	Gift
	PersonName string `json:"person_name"`
}

type ListParams struct {
	PersonID  *int64
	Direction Direction
	DebtType  DebtType
	PageSize  int
	Page      int
}

type GiftList struct {
	Items    []GiftWithPerson `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (g Gift) HasImage() bool { return g.ImagePath != "" }

func (g Gift) IsMoney() bool {
	return g.DebtType != DebtNone || g.AmountCents != nil
}

func (g Gift) DisplayAmount() string {
	if g.AmountCents == nil {
		return ""
	}

	major := *g.AmountCents / 100

	minor := *g.AmountCents % 100
	if minor == 0 {
		return fmt.Sprintf("%s %d", g.Currency, major)
	}

	return fmt.Sprintf("%s %d.%02d", g.Currency, major, minor)
}

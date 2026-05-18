package gifts

import (
	"fmt"
	"time"
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
	ID            int64     `json:"id"`
	PersonID      int64     `json:"person_id"`
	Title         string    `json:"title"`
	Direction     Direction `json:"direction"`
	Date          string    `json:"date"`
	Notes         string    `json:"notes"`
	AmountCents   *int64    `json:"amount_cents"`
	Currency      string    `json:"currency"`
	DebtType      DebtType  `json:"debt_type"`
	ImagePath     string    `json:"image_path"`
	ImageMimeType string    `json:"image_mime_type"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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

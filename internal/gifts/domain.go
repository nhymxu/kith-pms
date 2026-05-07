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
	ID            int64
	PersonID      int64
	Title         string
	Direction     Direction
	Date          string
	Notes         string
	AmountCents   *int64
	Currency      string
	DebtType      DebtType
	ImagePath     string
	ImageMimeType string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type GiftWithPerson struct {
	Gift
	PersonName string
}

type ListParams struct {
	PersonID  *int64
	Direction Direction
	DebtType  DebtType
	PageSize  int
	Page      int
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

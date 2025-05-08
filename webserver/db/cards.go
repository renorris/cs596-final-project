package db

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

type Card struct {
	UUID           uuid.UUID
	CreatedAt      time.Time
	FriendlyName   string
	RemainingOpens int
}

func (p *Pool) CreateCard(ctx context.Context, cardUUID uuid.UUID) (card *Card, err error) {
	card = &Card{
		UUID:           cardUUID,
		CreatedAt:      time.Now().UTC(),
		FriendlyName:   "New Card",
		RemainingOpens: 0,
	}

	if _, err = p.Exec(ctx, `
		INSERT INTO cards
		(uuid, created_at, friendly_name, remaining_opens)
		VALUES ($1, $2, $3, $4);`,
		card.UUID,
		card.CreatedAt,
		card.FriendlyName,
		card.RemainingOpens,
	); err != nil {
		return
	}

	return
}

func (p *Pool) ListCards(ctx context.Context) (cards []*Card, err error) {
	rows, err := p.Query(ctx, `
		SELECT 
		uuid, created_at, friendly_name, remaining_opens
		FROM cards
		ORDER BY created_at DESC;`,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	cards = make([]*Card, 0, 16)
	for rows.Next() {
		card := &Card{}
		if err = rows.Scan(
			&card.UUID,
			&card.CreatedAt,
			&card.FriendlyName,
			&card.RemainingOpens,
		); err != nil {
			return
		}

		cards = append(cards, card)
	}

	return
}

var NoMoreRemainingOpensError = errors.New("no more remaining opens")

// UseCard attempts to use a card. If the remaining_opens field for a Card
// is 0 or does not have infinite opens (-1), err will be non-nil.
func (p *Pool) UseCard(ctx context.Context, cardUUID uuid.UUID) (err error) {
	tx, err := p.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT remaining_opens FROM cards WHERE uuid = $1`,
		cardUUID,
	)

	var remainingOpens int64
	if err = row.Scan(&remainingOpens); err != nil {
		return
	}

	if remainingOpens == 0 {
		err = NoMoreRemainingOpensError
		return
	}

	// -1 indicates infinite opens
	if remainingOpens > 0 {
		if _, err = tx.Exec(ctx, `
			UPDATE cards
			SET remaining_opens = remaining_opens - 1 
			WHERE uuid = $1`,
			cardUUID,
		); err != nil {
			return
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	return
}

// SetCardOpens sets the number of allowed opens for card.
// Set to -1 for infinite opens, or 0 to disable.
func (p *Pool) SetCardOpens(ctx context.Context, cardUUID uuid.UUID, numOpens int) (err error) {
	if _, err = p.Exec(ctx, `
		UPDATE cards
		SET remaining_opens = $1
		WHERE uuid = $2;`, numOpens, cardUUID,
	); err != nil {
		return
	}

	return
}

// AddCardOpens adds a number of allowed opens for a card.
func (p *Pool) AddCardOpens(ctx context.Context, cardUUID uuid.UUID, numOpens int) (err error) {
	if _, err = p.Exec(ctx, `
		UPDATE cards
		SET remaining_opens = remaining_opens + $1
		WHERE uuid = $2`, numOpens, cardUUID,
	); err != nil {
		return
	}

	return
}

func (p *Pool) UpdateCardFriendlyName(ctx context.Context, cardUUID uuid.UUID, friendlyName string) (err error) {
	if _, err = p.Exec(ctx, `
		UPDATE cards
		SET friendly_name = $1
		WHERE uuid = $2;`, friendlyName, cardUUID,
	); err != nil {
		return
	}

	return
}

package storage

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID             uuid.UUID      `db:"id"`
	TelegramID     int64          `db:"telegram_id"`
	ChannelID      uuid.UUID      `db:"channel_id"`
	Text           sql.NullString `db:"text"`
	SentAt         time.Time      `db:"sent_at"`
	SenderUsername sql.NullString `db:"sender_username"`
	IsForward      sql.NullBool   `db:"is_forward"`
	MessageType    sql.NullString `db:"message_type"`
	RawData        []byte         `db:"raw_data"` // JSONB field
	CreatedAt      time.Time      `db:"created_at"`
}

// Prediction представляет структуру для хранения предсказаний в базе данных.
type Prediction struct {
	ID                  uuid.UUID      `db:"id"`
	MessageID           uuid.UUID      `db:"message_id"`
	Ticker              string         `db:"ticker"`
	PredictionType      sql.NullString `db:"prediction_type"`
	TargetPrice         sql.NullString `db:"target_price"`
	TargetChangePercent sql.NullString `db:"target_change_percent"`
	Period              sql.NullString `db:"period"`
	Recommendation      sql.NullString `db:"recommendation"`
	Direction           sql.NullString `db:"direction"`
	JustificationText   sql.NullString `db:"justification_text"`
	PredictedAt         time.Time      `db:"predicted_at"`
}

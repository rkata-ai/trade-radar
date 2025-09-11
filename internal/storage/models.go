package storage

import (
	"database/sql"
	"time"
)

type Message struct {
	ID             int64          `db:"id"`
	TelegramID     int64          `db:"telegram_id"`
	ChannelID      int64          `db:"channel_id"`
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
	ID                  int64           `db:"id"`
	MessageID           int64           `db:"message_id"`
	StockID             int64           `db:"stock_id"`
	PredictionType      sql.NullString  `db:"prediction_type"`
	TargetPrice         sql.NullFloat64 `db:"target_price"`
	TargetChangePercent sql.NullFloat64 `db:"target_change_percent"`
	Period              sql.NullString  `db:"period"`
	Recommendation      sql.NullString  `db:"recommendation"`
	Direction           sql.NullString  `db:"direction"`
	JustificationText   sql.NullString  `db:"justification_text"`
	PredictedAt         time.Time       `db:"predicted_at"`
}

type Industry struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

type Stock struct {
	ID         int64          `db:"id"`
	Ticker     string         `db:"ticker"`
	Name       sql.NullString `db:"name"`
	IndustryID sql.NullInt64  `db:"industry_id"`
	Exchange   sql.NullString `db:"exchange"`
	Currency   sql.NullString `db:"currency"`
	CreatedAt  time.Time      `db:"created_at"`
}

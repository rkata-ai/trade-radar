package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"rkata-ai/trade-radar/internal/config"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresStorage реализует интерфейс Storage для PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(cfg *config.DatabaseConfig) (Storage, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	if cfg.ConnectionString != "" {
		connStr = cfg.ConnectionString
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresStorage{
		db: db,
	}, nil
}

func (p *PostgresStorage) GetMessagesWithoutPredictions(ctx context.Context, limit int) ([]Message, error) {
	const op = "storage.GetMessagesWithoutPredictions"

	messages := []Message{}
	query := `
		SELECT
			m.id, m.telegram_id, m.channel_id, m.text, m.sent_at, m.sender_username, m.is_forward, m.message_type, m.raw_data, m.created_at
		FROM
			messages m
		LEFT JOIN
			predictions p ON m.id = p.message_id
		WHERE
			p.id IS NULL
		ORDER BY
			m.sent_at ASC
		LIMIT $1
	`
	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages without predictions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		message := Message{}
		err := rows.Scan(
			&message.ID,
			&message.TelegramID,
			&message.ChannelID,
			&message.Text,
			&message.SentAt,
			&message.SenderUsername,
			&message.IsForward,
			&message.MessageType,
			&message.RawData,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", op, err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", op, err)
	}

	return messages, nil
}

func (p *PostgresStorage) SavePrediction(ctx context.Context, prediction *Prediction) error {
	const op = "storage.SavePrediction"

	query := `
		INSERT INTO predictions (
			id, message_id, ticker, prediction_type, target_price, 
			target_change_percent, period, recommendation, direction, 
			justification_text, predicted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id
	`

	var id uuid.UUID
	err := p.db.QueryRowContext(ctx, query,
		prediction.ID,
		prediction.MessageID,
		prediction.Ticker,
		prediction.PredictionType,
		prediction.TargetPrice,
		prediction.TargetChangePercent,
		prediction.Period,
		prediction.Recommendation,
		prediction.Direction,
		prediction.JustificationText,
		prediction.PredictedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to save prediction: %w", op, err)
	}

	return nil
}

func (p *PostgresStorage) Close() error {
	const op = "storage.Close"
	return p.db.Close()
}

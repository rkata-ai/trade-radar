package storage

import (
	"context"
)

// Storage определяет интерфейс для взаимодействия с базой данных
type Storage interface {
	GetMessagesWithoutPredictions(ctx context.Context, limit int) ([]Message, error)
	SavePrediction(ctx context.Context, prediction *Prediction) error
	GetStock(ctx context.Context, ticker string) (*Stock, error)
	SaveRawPrediction(ctx context.Context, rawPrediction *RawPrediction) error
	Close() error
}

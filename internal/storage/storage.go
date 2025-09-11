package storage

import (
	"context"
)

// Storage определяет интерфейс для взаимодействия с базой данных
type Storage interface {
	GetMessagesWithoutPredictions(ctx context.Context, limit int) ([]Message, error)
	SavePrediction(ctx context.Context, prediction *Prediction) error
	GetOrCreateStock(ctx context.Context, ticker string) (*Stock, error)
	Close() error
}

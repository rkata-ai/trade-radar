package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"rkata-ai/trade-radar/internal/ai"
	"rkata-ai/trade-radar/internal/config"
	"rkata-ai/trade-radar/internal/storage"

	"github.com/google/uuid"
)

func main() {
	// Парсинг флагов командной строки
	var configPath string
	var outputTo string
	var outputFilePath string
	var debugFlag bool

	flag.StringVar(&configPath, "config", "", "Path to configuration file (required)")
	flag.StringVar(&outputTo, "output-to", "console", "Output results to 'console' or 'file' or 'db'")
	flag.StringVar(&outputFilePath, "output-file", "analysis_results.json", "Path to the output JSON file if output-to is 'file'")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging, including raw Ollama responses")
	flag.Parse()

	outputTo = strings.TrimSpace(outputTo) // Очищаем значение outputTo от пробельных символов

	// Инициализация логгера
	logger := log.Default()

	// Проверяем, что флаг config передан
	if configPath == "" {
		logger.Printf("Usage: ./bin/trading.exe -config <path_to_config> [--output-to <console|file|db>] [--output-file <path>]")
		os.Exit(1)
	}

	logger.Print("Starting R&D research for trading channel rating system")
	logger.Printf("Using config file: %s", configPath)

	// Загрузка конфигурации
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Логируем загруженные параметры конфигурации
	logger.Printf("Config loaded successfully:")
	logger.Printf("  AI.OllamaBaseURL: %s", cfg.AI.OllamaBaseURL)
	logger.Printf("  AI.OllamaModel: %s", cfg.AI.OllamaModel)
	logger.Printf("  AI.Debug: %t", cfg.AI.Debug)
	logger.Printf("  Database.Host: %s", cfg.Database.Host)
	logger.Printf("  Database.Port: %d", cfg.Database.Port)
	logger.Printf("  Database.User: %s", cfg.Database.User)
	logger.Printf("  Database.DBName: %s", cfg.Database.DBName)
	logger.Printf("  Database.SSLMode: %s", cfg.Database.SSLMode)
	logger.Printf("  Database.ConnectionString: %s", cfg.Database.ConnectionString)

	// Инициализация AI клиента
	aiClient := ai.NewOllamaClient(cfg.AI.OllamaBaseURL, cfg.AI.OllamaModel, debugFlag)

	// Инициализация хранилища базы данных
	var dbStorage storage.Storage
	dbStorage, err = storage.NewPostgresStorage(&cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbStorage.Close()

	var messages []storage.Message
	messages, err = dbStorage.GetMessagesWithoutPredictions(context.Background(), 100) // Ограничиваем до 100 сообщений за раз
	if err != nil {
		logger.Fatalf("Failed to get messages from database: %v", err)
	}

	if len(messages) == 0 {
		logger.Print("No new messages to analyze.")
		return
	}
	logger.Printf("Read %d messages from database", len(messages))

	var allAnalyses []*ai.MessageAnalysis
	for idx, message := range messages {
		logger.Printf("Analyzing message %d/%d (ID: %s): %s", idx+1, len(messages), message.ID.String(), message.Text.String)
		analysis, err := aiClient.AnalyzeMessage(context.Background(), message.Text.String, message.ChannelID.String(), message.ID)
		if err != nil {
			logger.Printf("Failed to analyze message %d (ID: %s): %v", idx+1, message.ID.String(), err)
			continue
		}
		allAnalyses = append(allAnalyses, analysis)

		if outputTo == "db" {
			// Сохраняем прогнозы в базу данных
			for _, pred := range analysis.Predictions {
				// Проверяем, что Ticker не пустой и PredictionType не 'Неопределенный' перед сохранением
				if pred.Ticker == "" || pred.PredictionType == "Неопределенный" {
					logger.Printf("Prediction for message %s with ticker '%s' and type '%s' ignored (empty ticker or 'Неопределенный' type). Skipping.", message.ID.String(), pred.Ticker, pred.PredictionType)
					continue
				}

				dbPrediction := storage.Prediction{
					ID:                  uuid.New(),
					MessageID:           message.ID,
					Ticker:              pred.Ticker,
					PredictionType:      sql.NullString{String: pred.PredictionType, Valid: pred.PredictionType != ""},
					TargetPrice:         sql.NullString{String: pred.TargetPrice.String(), Valid: !pred.TargetPrice.IsNull},
					TargetChangePercent: sql.NullString{String: pred.TargetChangePercent.String(), Valid: !pred.TargetChangePercent.IsNull},
					Period:              sql.NullString{String: pred.Period, Valid: pred.Period != ""},
					Recommendation:      sql.NullString{String: pred.Recommendation, Valid: pred.Recommendation != ""},
					Direction:           sql.NullString{String: pred.Direction, Valid: pred.Direction != ""},
					JustificationText:   sql.NullString{String: pred.JustificationText, Valid: pred.JustificationText != ""},
					PredictedAt:         time.Now(),
				}

				err = dbStorage.SavePrediction(context.Background(), &dbPrediction)
				if err != nil {
					logger.Printf("Failed to save prediction for message %s: %v", message.ID.String(), err)
				} else {
					logger.Printf("Prediction for message %s saved to DB.", message.ID.String())
				}
			}
		} else {
			switch outputTo {
			case "console":
				logger.Printf("\n### Message %d ###\n", idx+1)
				if len(analysis.Predictions) > 0 {
					for i, prediction := range analysis.Predictions {
						// Проверяем, что Ticker не пустой и PredictionType не 'Неопределенный' перед выводом в консоль
						if prediction.Ticker == "" || prediction.PredictionType == "Неопределенный" {
							logger.Printf("Prediction for message %s with ticker '%s' and type '%s' ignored (empty ticker or 'Неопределенный' type). Skipping console output.", prediction.MessageID.String(), prediction.Ticker, prediction.PredictionType)
							continue
						}
						logger.Printf("--- Prediction %d ---", i+1)
						logger.Printf("  Message ID: %s", prediction.MessageID.String())
						logger.Printf("  Prediction Type: %s", prediction.PredictionType)
						logger.Printf("  Ticker: %s", prediction.Ticker)
						logger.Printf("  Target Price: %s", prediction.TargetPrice.String())
						logger.Printf("  Target Change Percent: %s", prediction.TargetChangePercent.String())
						logger.Printf("  Period: %s", prediction.Period)
						logger.Printf("  Recommendation: %s", prediction.Recommendation)
						logger.Printf("  Direction: %s", prediction.Direction)
						logger.Printf("  Justification Text: %s\n", prediction.JustificationText)
					}
				} else {
					logger.Printf("  No financial predictions available for message %d.\n", idx+1)
				}
			case "file":
				var filteredPredictions []ai.FinancialPrediction
				for _, pred := range analysis.Predictions {
					if pred.Ticker != "" && pred.PredictionType != "Неопределенный" {
						filteredPredictions = append(filteredPredictions, pred)
					}
				}
				// Если есть валидные предсказания, создаем новый MessageAnalysis и добавляем его в allAnalyses
				if len(filteredPredictions) > 0 {
					filteredAnalysis := &ai.MessageAnalysis{
						Predictions: filteredPredictions,
					}
					allAnalyses = append(allAnalyses, filteredAnalysis)
				}

				outputJSON, marshalErr := json.MarshalIndent(allAnalyses, "", "  ")
				if marshalErr != nil {
					logger.Printf("Failed to marshal analysis results to JSON for message %d: %v", idx+1, marshalErr)
				} else {
					err := os.WriteFile(outputFilePath, outputJSON, 0644)
					if err != nil {
						logger.Printf("Failed to write analysis results to file %s for message %d: %v", outputFilePath, idx+1, err)
					} else {
						logger.Printf("Analysis results for message %d written to %s", idx+1, outputFilePath)
					}
				}
			}
		}
	}

	// Обработка случая, если outputTo не является ни 'console', ни 'file', ни 'db'
	if outputTo != "console" && outputTo != "file" && outputTo != "db" {
		logger.Fatalf("Invalid output-to option: %s. Use 'console', 'file' or 'db'.", outputTo)
	}

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала завершения
	logger.Print("\n=== Processing completed successfully! ===")
	logger.Print("Press Ctrl+C to exit...")
	<-sigChan
	logger.Print("Shutting down...")
}

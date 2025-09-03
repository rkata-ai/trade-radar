package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"traiding/internal/ai"
	"traiding/internal/config"
)

func main() {
	// Парсинг флагов командной строки
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file (required)")
	flag.Parse()

	// Инициализация логгера
	logger := log.Default()

	// Проверяем, что флаг config передан
	if configPath == "" {
		logger.Printf("Usage: ./bin/trading.exe -config <path_to_config>")
		logger.Printf("Example: ./bin/trading.exe -config bin/config.yaml")
		os.Exit(1)
	}

	logger.Print("Starting R&D research for trading channel rating system")
	logger.Printf("Using config file: %s", configPath)

	// Загрузка конфигурации
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация AI клиента
	aiClient := ai.NewOllamaClient(cfg.AI.OllamaBaseURL, cfg.AI.OllamaModel)

	// Пример использования AnalyzeMessage
	testMessage := "Акция XYZ демонстрирует сильный восходящий тренд, рекомендуем покупать с целью 150 долларов в краткосрочной перспективе."
	testChannel := "TestChannel"

	logger.Printf("Analyzing message: %s", testMessage)
	analysis, err := aiClient.AnalyzeMessage(context.Background(), testMessage, testChannel)
	if err != nil {
		logger.Fatalf("Failed to analyze message: %v", err)
	}

	logger.Print("\n--- AI Analysis (Ollama): ---")
	logger.Printf("  Summary: %s", analysis.Summary)

	if len(analysis.Predictions) > 0 {
		for i, prediction := range analysis.Predictions {
			logger.Printf("\n--- Prediction %d ---", i+1)
			logger.Printf("  Message ID: %s", prediction.MessageID)
			logger.Printf("  Prediction Type: %s", prediction.PredictionType)
			logger.Printf("  Ticker: %s", prediction.Ticker)
			logger.Printf("  Target Price: %s", prediction.TargetPrice)
			logger.Printf("  Target Change Percent: %.2f", prediction.TargetChangePercent)
			logger.Printf("  Period: %s", prediction.Period)
			logger.Printf("  Full Text: %s", prediction.FullText)
			logger.Printf("  Recommendation: %s", prediction.Recommendation)
		}
	} else {
		logger.Print("  No financial predictions available.")
	}

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала завершения
	logger.Print("\n=== Test completed successfully! ===")
	logger.Print("Press Ctrl+C to exit...")
	<-sigChan
	logger.Print("Shutting down...")
}

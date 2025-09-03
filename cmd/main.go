package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"traiding/internal/ai"
	"traiding/internal/config"
)

func main() {
	// Парсинг флагов командной строки
	var configPath string
	var inputFilePath string
	var outputTo string
	var outputFilePath string
	var debugFlag bool

	flag.StringVar(&configPath, "config", "", "Path to configuration file (required)")
	flag.StringVar(&inputFilePath, "input-file", "", "Path to a text file with messages (one message per line, required)")
	flag.StringVar(&outputTo, "output-to", "console", "Output results to 'console' or 'file'")
	flag.StringVar(&outputFilePath, "output-file", "analysis_results.json", "Path to the output JSON file if output-to is 'file'")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging, including raw Ollama responses")
	flag.Parse()

	outputTo = strings.TrimSpace(outputTo) // Очищаем значение outputTo от пробельных символов

	// Инициализация логгера
	logger := log.Default()

	// Проверяем, что флаги config и input-file переданы
	if configPath == "" || inputFilePath == "" {
		logger.Printf("Usage: ./bin/trading.exe -config <path_to_config> -input-file <path> [--output-to <console|file>] [--output-file <path>]")
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
	aiClient := ai.NewOllamaClient(cfg.AI.OllamaBaseURL, cfg.AI.OllamaModel, debugFlag)

	var messages []string
	testChannel := "TestChannel"

	file, err := os.Open(inputFilePath)
	if err != nil {
		logger.Fatalf("Failed to open input file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		message := scanner.Text()
		if message != "" {
			messages = append(messages, message)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Fatalf("Failed to read input file: %v", err)
	}
	if len(messages) == 0 {
		logger.Fatal("Input file is empty or contains no messages.")
	}
	logger.Printf("Read %d messages from %s", len(messages), inputFilePath)

	var allAnalyses []*ai.MessageAnalysis
	for idx, message := range messages {
		logger.Printf("Analyzing message %d/%d: %s", idx+1, len(messages), message)
		analysis, err := aiClient.AnalyzeMessage(context.Background(), message, testChannel)
		if err != nil {
			logger.Printf("Failed to analyze message %d: %v", idx+1, err)
			continue
		}
		allAnalyses = append(allAnalyses, analysis) // Продолжаем собирать результаты для вывода в файл

		switch outputTo {
		case "console":
			logger.Printf("\n### Message %d ###\n", idx+1)
			if len(analysis.Predictions) > 0 {
				for i, prediction := range analysis.Predictions {
					logger.Printf("--- Prediction %d ---", i+1)
					logger.Printf("  Message ID: %s", prediction.MessageID)
					logger.Printf("  Prediction Type: %s", prediction.PredictionType)
					logger.Printf("  Ticker: %s", prediction.Ticker)
					logger.Printf("  Target Price: %s", prediction.TargetPrice)
					logger.Printf("  Target Change Percent: %s", prediction.TargetChangePercent)
					logger.Printf("  Period: %s", prediction.Period)
					logger.Printf("  Recommendation: %s", prediction.Recommendation)
					logger.Printf("  Direction: %s", prediction.Direction)
					logger.Printf("  Justification Text: %s\n", prediction.JustificationText)
				}
			} else {
				logger.Printf("  No financial predictions available for message %d.\n", idx+1)
			}
		case "file":
			// Записываем текущий батч в файл после каждого сообщения
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

	// Обработка случая, если outputTo не является ни 'console', ни 'file'
	if outputTo != "console" && outputTo != "file" {
		logger.Fatalf("Invalid output-to option: %s. Use 'console' or 'file'.", outputTo)
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

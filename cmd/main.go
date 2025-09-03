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
	testMessage := "#Аэрофлот (#AFLT) - прогноз реализовался, что дальше?  Пару недель назад писал, что акции авиаперевозчика - одни из главных кандидатов на отскок. Технически тогда всё выглядело максимально лонгово 👇  1️⃣ Отскок котировок от сильнейшего уровня 46 и глобальной трендовой, которая начала формироваться сразу после мобилизации  2️⃣ Сформированный разворотный паттерн доджи на важнейших значениях из 1 пункта.   Да и в целом цена пришла на точку входа, от которой в сентябре показала 30% рост. Так что получили сильнейшее комбо - со своих минимумов акции выросли также на 30%.  Скептики часто любят говорить - самолетов нет у нас, всё ломается, долг большой. Откуда ж взяться росту?  Как видите, такая банальная логика совершенно не работает. Это из разряда мыслей новичков, что Белуга и Абрау Дюрсо непременно сильно вырастут перед НГ (все ж алкашку покупают), а акции Whoosh обязательно будут падать с ноябрьского окончания сезона.  Реально всё работает иначе - у Аэрофлота растёт пассажиропоток и цены на авиабилеты, для инвесторов это куда важнее остальных проблем. Именно поэтому в 3 квартале сильно выросла выручка (+37%), чистая прибыль составила 17.6 млрд против 9.3 млрд убытка годом ранее и сократился чистый долг на 9.7%.  Что по технике? В районе 64 очень сильный уровень, от которого начиналась сильная коррекция в июне и ноябре. До сопротивления порядка 7-8%, поэтому небольшое пространство для роста ещё есть.  Но пытаться выжать от позиции по максимуму далеко не всегда заканчивается полезно для портфеля, поэтому начинать фиксироваться надо уже сейчас, благо прибыль весьма жирная (от 20% до 25% в зависимости от точки входа).  А вот при штурме сопротивления 64 и попытках отката можно будет присмотреться к шорту. Но пока об этом рано, если будут технические предпосылки для коррекции, напишу в канале 😎"
	testChannel := "TestChannel"

	logger.Printf("Analyzing message: %s", testMessage)
	analysis, err := aiClient.AnalyzeMessage(context.Background(), testMessage, testChannel)
	if err != nil {
		logger.Fatalf("Failed to analyze message: %v", err)
	}

	logger.Print("\n--- AI Analysis (Ollama): ---")
	if len(analysis.Predictions) > 0 {
		for i, prediction := range analysis.Predictions {
			logger.Printf("\n--- Prediction %d ---", i+1)
			logger.Printf("  Message ID: %s", prediction.MessageID)
			logger.Printf("  Prediction Type: %s", prediction.PredictionType)
			logger.Printf("  Ticker: %s", prediction.Ticker)
			logger.Printf("  Target Price: %s", prediction.TargetPrice)
			logger.Printf("  Target Change Percent: %s", prediction.TargetChangePercent)
			logger.Printf("  Period: %s", prediction.Period)
			logger.Printf("  Recommendation: %s", prediction.Recommendation)
			logger.Printf("  Direction: %s", prediction.Direction)
			logger.Printf("  Justification Text: %s", prediction.JustificationText)
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

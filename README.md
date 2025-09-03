# Система анализа финансовых рекомендаций

R&D исследование для создания системы оценки эффективности торговых рекомендаций.

## 🎯 Цель проекта

Создать систему, которая:
1. Принимает текстовые сообщения с финансовыми рекомендациями.
2. Анализирует их с помощью AI-модели Ollama.
3. Извлекает структурированные финансовые прогнозы.

## 🏗️ Архитектура

```
traiding/
├── cmd/main.go              # Главный файл для R&D исследования
├── internal/
│   ├── ai/                  # AI-клиент Ollama и логика анализа
│   └── config/              # Конфигурация
├── configs/                 # Конфигурационные файлы
└── go.mod                   # Зависимости Go
```

## 🚀 Быстрый старт

### Предварительные требования

1. **Go 1.21+** - [скачать](https://golang.org/dl/)
2. **Ollama** - установленный локальный сервер Ollama с загруженной моделью (например, `gemma3:1b`, но выбор зависит от мощности вашего компьютера и желаемой скорости обработки). [Скачать Ollama](https://ollama.com/)

### Установка зависимостей

```bash
go mod tidy
```

### Настройка конфигурации

1. Скопируйте и отредактируйте конфигурационный файл:
```bash
cp configs/config.yaml configs/config.local.yaml
```

2. Заполните необходимые параметры (пример для `configs/config.local.yaml`):
```yaml
ai:
  ollama_base_url: "http://localhost:11434" # URL вашего сервера Ollama
  ollama_model: "gemma3:1b"               # Используемая модель Ollama
```

### Запуск R&D исследования

```bash
# Запуск с указанием конфигурационного файла (ОБЯЗАТЕЛЬНО)
go run cmd/main.go -config configs/config.local.yaml

# Запуск собранной программы
./bin/traiding.exe -config configs/config.local.yaml

# Показать справку по флагам
./bin/traiding.exe -help
```

## 🔧 Конфигурация

### Флаги командной строки

```bash
# Основные флаги
-config string    # Путь к конфигурационному файлу (ОБЯЗАТЕЛЬНО)
-help             # Показать справку по флагам
```

### Структура конфигурации

```yaml
ai:
  ollama_base_url: "http://localhost:11434" # URL вашего сервера Ollama
  ollama_model: "gemma3:1b"               # Используемая модель Ollama
```

## 📊 Что тестируется

- [x] Базовая структура проекта
- [x] Конфигурация и логирование
- [x] Ollama AI-клиент
- [x] Извлечение структурированных финансовых прогнозов с помощью Ollama

## 🔄 Расширение AI-пайплайна анализа

Система анализа сообщений реализована с использованием паттерна "пайплайн", что позволяет легко добавлять, изменять или удалять шаги обработки без затрагивания основной логики.

### Интерфейс PipelineStep

Все шаги пайплайна должны реализовывать интерфейс `PipelineStep`, определенный в `internal/ai/ollama_client.go`:

```go
type PipelineStep interface {
	Execute(ctx context.Context, client *OllamaClient, message string, messageID string) ([]FinancialPrediction, error)
}
```

-   `Execute`: Метод, который выполняет логику конкретного шага. Он принимает контекст, клиент Ollama, исходное сообщение и уникальный идентификатор сообщения, возвращая срез финансовых прогнозов.

### Создание нового шага

Для добавления нового шага в пайплайн:

1.  **Создайте новую структуру**, которая будет представлять ваш шаг (например, `SentimentAnalysisStep`).
2.  **Реализуйте интерфейс `PipelineStep`** для этой структуры. Внутри метода `Execute` вы можете выполнять любую необходимую логику: вызывать Ollama с другим промтом, обрабатывать данные, фильтровать прогнозы и т.д.

Пример нового шага (для демонстрации):

```go
// SentimentAnalysisStep реализует PipelineStep для анализа настроения сообщения.
type SentimentAnalysisStep struct{}

func NewSentimentAnalysisStep() *SentimentAnalysisStep {
	return &SentimentAnalysisStep{}
}

func (s *SentimentAnalysisStep) Execute(ctx context.Context, client *OllamaClient, message string, messageID string) ([]FinancialPrediction, error) {
	// Пример логики: вызов Ollama для анализа настроения
	// prompt := fmt.Sprintf("Проанализируй настроение этого сообщения: %s", message)
	// res, err := client.sendOllamaRequest(ctx, prompt)
	// ... обработка ответа ...

	// Для простоты возвращаем пустой срез, но здесь была бы реальная логика.
	return []FinancialPrediction{}, nil
}
```

### Интеграция шагов в пайплайн

На данный момент функция `AnalyzeMessage` в `internal/ai/ollama_client.go` содержит логику для выполнения единственного шага `PredictionStep`:

```go
func (c *OllamaClient) AnalyzeMessage(ctx context.Context, message, channel string) (*MessageAnalysis, error) {
	messageID := "some_unique_id" // TODO: Заменить на реальную генерацию ID

	predictionStep := NewPredictionStep()
	predictions, err := predictionStep.Execute(ctx, c, message, messageID)
	// ... обработка ошибок ...

	return &MessageAnalysis{
		Predictions: predictions,
	}, nil
}
```

Для добавления нескольких шагов вы можете создать композицию шагов или массив `PipelineStep` и выполнять их последовательно, передавая результаты одного шага в качестве входных данных для следующего (если это требуется вашей логикой).

Например, можно создать массив шагов и выполнять их в цикле, или же модифицировать `PredictionStep` для включения других шагов.

```go
// Пример простого последовательного выполнения нескольких шагов
func (c *OllamaClient) AnalyzeMessage(ctx context.Context, message, channel string) (*MessageAnalysis, error) {
	messageID := "some_unique_id"

	steps := []ai.PipelineStep{
		ai.NewPredictionStep(),
		// ai.NewSentimentAnalysisStep(), // Добавьте ваш новый шаг здесь
	}

	var allPredictions []ai.FinancialPrediction
	for _, step := range steps {
		predictions, err := step.Execute(ctx, c, message, messageID)
		if err != nil {
			return nil, fmt.Errorf("failed to execute pipeline step: %w", err)
		}
		allPredictions = append(allPredictions, predictions...)
	}

	return &ai.MessageAnalysis{
		Predictions: allPredictions,
	}, nil
}
```

## 🔮 Следующие шаги

1. **Расширение пайплайна анализа** - добавление новых шагов в обработку AI-анализа.
2. **Интеграция с источниками данных** - подключение к реальным источникам сообщений (например, Telegram) и историческим данным котировок (например, MetaTrader).
3. **База данных** - сохранение и анализ исторических данных.
4. **Алгоритмы расчета рейтинга** - разработка и внедрение алгоритмов для оценки эффективности рекомендаций.
5. **Веб-интерфейс** - создание пользовательского интерфейса для отображения результатов.

## 📝 Лицензия

MIT License

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для новой функции
3. Внесите изменения
4. Создайте Pull Request

## 📞 Поддержка

По вопросам и предложениям создавайте Issues в репозитории.


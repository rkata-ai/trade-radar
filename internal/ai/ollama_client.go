package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// FlexibleFloatOrString представляет собой поле, которое может быть либо float64, либо string.
type FlexibleFloatOrString struct {
	FloatValue  float64
	StringValue string
	IsFloat     bool
	IsNull      bool // Добавлено для отслеживания значения null
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для FlexibleFloatOrString.
func (ffos *FlexibleFloatOrString) UnmarshalJSON(data []byte) error {
	// Проверка на JSON null
	if string(data) == "null" {
		ffos.IsNull = true
		return nil
	}

	// Попытка демаршалинга как float64
	if err := json.Unmarshal(data, &ffos.FloatValue); err == nil {
		ffos.IsFloat = true
		ffos.IsNull = false
		return nil
	}

	// Если не float64, попытка демаршалинга как string
	if err := json.Unmarshal(data, &ffos.StringValue); err == nil {
		ffos.IsFloat = false
		ffos.IsNull = false
		return nil
	}

	return fmt.Errorf("could not unmarshal into float64 or string: %s", data)
}

// String возвращает строковое представление FlexibleFloatOrString.
func (ffos FlexibleFloatOrString) String() string {
	if ffos.IsNull {
		return ""
	}
	if ffos.IsFloat {
		return fmt.Sprintf("%.2f", ffos.FloatValue)
	}
	return ffos.StringValue
}

type MessageAnalysis struct {
	Predictions []FinancialPrediction
}

type FinancialPrediction struct {
	MessageID           string                `json:"message_id"`
	PredictionType      string                `json:"prediction_type"`
	Ticker              string                `json:"ticker"`
	TargetPrice         string                `json:"target_price"`
	TargetChangePercent FlexibleFloatOrString `json:"target_change_percent"` // Изменено на FlexibleFloatOrString
	Period              string                `json:"period"`
	Recommendation      string                `json:"recommendation"`
	Direction           string                `json:"direction"`
	JustificationText   string                `json:"justification_text"`
}

type AIClient interface {
	AnalyzeMessage(ctx context.Context, message, channel string) (*MessageAnalysis, error)
	AnalyzeBatch(ctx context.Context, messages []string, channel string) ([]*MessageAnalysis, error)
}

type OllamaClient struct {
	baseURL string
	model   string
}

// OllamaGenerateRequest - структура для запроса к Ollama API /api/generate
type OllamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaGenerateResponse - структура для ответа от Ollama API /api/generate
type OllamaGenerateResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"` // Основной текст ответа
	Done      bool   `json:"done"`
}

func NewOllamaClient(baseURL string, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
	}
}

// PipelineStep определяет интерфейс для шага в конвейере анализа сообщений.
type PipelineStep interface {
	Execute(ctx context.Context, client *OllamaClient, message string, messageID string) ([]FinancialPrediction, error)
}

// PredictionStep реализует PipelineStep для выполнения финансового прогнозирования.
type PredictionStep struct {
	// Здесь могут быть дополнительные поля, если шаг требует собственной конфигурации
}

// NewPredictionStep создает новый экземпляр PredictionStep.
func NewPredictionStep() *PredictionStep {
	return &PredictionStep{}
}

// Execute выполняет шаг прогнозирования, используя предоставленный промт.
func (s *PredictionStep) Execute(ctx context.Context, client *OllamaClient, message string, messageID string) ([]FinancialPrediction, error) {
	prompt := fmt.Sprintf(`
	Ты опытный финансовый аналитик. Твоя задача — извлечь из предоставленного сообщения прогнозы по акциям и структурировать их в формате JSON. Если какая-либо информация (например, целевая цена или период) отсутствует, используй значение null.
	Формат ответа: JSON-массив, содержащий один или несколько объектов. Каждый объект должен иметь следующие поля:
	prediction_type: Тип прогноза. Выбери наиболее подходящий вариант из списка: "Продолжение тренда", "Разворот", "Цель с коррекцией", "Накопление перед пробоем", "Долгосрочный пессимизм", "Неопределенный".
	ticker: Тикер акции, указанный в сообщении (например, AFLT).
	period: Временной горизонт прогноза. Используй один из вариантов: "Сегодня", "Краткосрочный", "Среднесрочный", "Долгосрочный", "Неопределенный".
	target_price: Целевая цена или ценовой диапазон. Извлеки числовое значение. Если в сообщении указан диапазон, используй строку, например "60-65". Если цена не указана, используй null.
	target_change_percent: Целевой процент изменения цены. Извлеки числовое значение или диапазон в виде строки, например "7-8%%". Если процент не указан, используй null.
	recommendation: Рекомендация автора сообщения. Используй один из вариантов: "Покупать", "Продавать", "Держать", "Неопределенный".
	direction: Направление сделки. Выбери один из вариантов: "Лонг", "Шорт", "Неопределенный".
	justification_text: Цитата из исходного текста, которая подтверждает данный прогноз.

	Сообщение: %s

	Отвечай только JSON, без дополнительного текста.`, message)

	req := OllamaGenerateRequest{
		Model:  client.model,
		Prompt: prompt,
		Stream: false,
	}

	res, err := client.sendOllamaRequest(ctx, req.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to send ollama request: %w", err)
	}

	var ollamaResponse OllamaGenerateResponse
	if err := json.Unmarshal(res, &ollamaResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response JSON: %w, body: %s", err, string(res))
	}

	content := ollamaResponse.Response
	content = strings.TrimSpace(content)

	// Пытаемся найти JSON в ответе (Ollama иногда добавляет перед/после JSON)
	jsonStart := strings.Index(content, "[")
	jsonEnd := strings.LastIndex(content, "]")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("invalid JSON response from Ollama: %s", content)
	}

	jsonContent := content[jsonStart : jsonEnd+1]

	// log.Printf("Ollama Financial Prediction Raw JSON Content: %s", jsonContent) // Логируем извлеченный JSON

	var predictions []FinancialPrediction
	if err := json.Unmarshal([]byte(jsonContent), &predictions); err != nil {
		// Если не удалось демаршалировать как слайс, пробуем как одиночный объект
		var singlePrediction FinancialPrediction
		if err := json.Unmarshal([]byte(jsonContent), &singlePrediction); err == nil {
			predictions = []FinancialPrediction{singlePrediction}
		} else {
			return nil, fmt.Errorf("failed to unmarshal financial prediction JSON: %w, content: %s", err, jsonContent)
		}
	}

	return predictions, nil
}

// sendOllamaRequest отправляет запрос к Ollama API и возвращает байты ответа.
func (c *OllamaClient) sendOllamaRequest(ctx context.Context, prompt string) ([]byte, error) {
	requestBody := OllamaGenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false, // Мы хотим получить весь ответ сразу
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama api returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ollama response body: %w", err)
	}

	return bodyBytes, nil
}

func (c *OllamaClient) AnalyzeMessage(ctx context.Context, message, channel string) (*MessageAnalysis, error) {
	messageID := "some_unique_id" // TODO: Заменить на реальную генерацию ID

	predictionStep := NewPredictionStep()
	predictions, err := predictionStep.Execute(ctx, c, message, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute prediction step: %w", err)
	}

	return &MessageAnalysis{
		Predictions: predictions,
	}, nil
}

func (c *OllamaClient) AnalyzeBatch(ctx context.Context, messages []string, channel string) ([]*MessageAnalysis, error) {
	var analyses []*MessageAnalysis

	for _, message := range messages {
		analysis, err := c.AnalyzeMessage(ctx, message, channel)
		if err != nil {
			fmt.Printf("Failed to analyze message: %v\n", err)
			continue
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

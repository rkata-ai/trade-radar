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

type MessageAnalysis struct {
	Summary     string                `json:"summary"`
	Predictions []FinancialPrediction `json:"predictions"` // Теперь слайс прогнозов
}

type FinancialPrediction struct {
	MessageID           string  `json:"message_id"`
	PredictionType      string  `json:"prediction_type"`
	Ticker              string  `json:"ticker"`
	TargetPrice         string  `json:"target_price"`
	TargetChangePercent float64 `json:"target_change_percent"`
	Period              string  `json:"period"`
	FullText            string  `json:"full_text"`
	Recommendation      string  `json:"recommendation"`
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

// SummaryPromptRequest - структура для запроса резюме
type SummaryPromptRequest struct {
	Message string `json:"message"`
}

// SummaryPromptResponse - структура для ответа резюме
type SummaryPromptResponse struct {
	Summary string `json:"summary"`
}

// FinancialPredictionPromptRequest - структура для запроса финансового прогноза
type FinancialPredictionPromptRequest struct {
	Summary string `json:"summary"`
	Channel string `json:"channel"`
}

// FinancialPredictionPromptResponse - структура для ответа финансового прогноза
type FinancialPredictionPromptResponse []FinancialPrediction // Теперь массив FinancialPrediction

func NewOllamaClient(baseURL string, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
	}
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
		return nil, fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Ollama response body: %w", err)
	}

	return bodyBytes, nil
}

// parseOllamaResponse парсит JSON-ответ от Ollama и извлекает содержимое.
func (c *OllamaClient) parseOllamaResponse(bodyBytes []byte) (string, error) {
	var ollamaResponse OllamaGenerateResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResponse); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response JSON: %w, body: %s", err, string(bodyBytes))
	}

	content := ollamaResponse.Response
	content = strings.TrimSpace(content)

	// Пытаемся найти JSON в ответе (Ollama иногда добавляет перед/после JSON)
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return "", fmt.Errorf("invalid JSON response from Ollama: %s", content)
	}

	jsonContent := content[jsonStart : jsonEnd+1]
	return jsonContent, nil
}

// SummarizeMessage отправляет сообщение в Ollama для получения резюме.
func (c *OllamaClient) SummarizeMessage(ctx context.Context, message string) (string, error) {
	prompt := fmt.Sprintf(`
	Резюмируй сообщение, сохраняй величины значений, подтверждающие выжимку

	Сообщение: %s

	Ответь только резюме в JSON формате:
	{
	    "summary": "Краткое резюме на русском языке"
	}
	`, message)

	bodyBytes, err := c.sendOllamaRequest(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to send Ollama summary request: %w", err)
	}

	jsonContent, err := c.parseOllamaResponse(bodyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse Ollama summary response: %w", err)
	}

	var summaryResponse SummaryPromptResponse
	if err := json.Unmarshal([]byte(jsonContent), &summaryResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal summary JSON: %w, content: %s", err, jsonContent)
	}

	return summaryResponse.Summary, nil
}

// PredictFinancialOutcome отправляет резюме и канал в Ollama для получения финансового прогноза.
func (c *OllamaClient) PredictFinancialOutcome(ctx context.Context, summary, channel, messageID string) (*FinancialPredictionPromptResponse, error) {
	prompt := fmt.Sprintf(`
	На основе выжимки извлеки из текста финансовые прогнозы. Используй предоставленные категории для 'prediction_type' и 'period'. Если информация отсутствует, укажи 'null' или 'Неопределенный'. Также дай рекомендацию если она есть. Результат представь в виде JSON массива.

	Резюме сообщения: %s

	Шаблон ответа в JSON:

	[
	{
	"message_id": "%s",
	"prediction_type": "Продолжение тренда" | "Разворот" | "Цель с коррекцией" | "Накопление перед пробоем" | "Долгосрочный пессимизм" | "Неопределенный",
	"ticker": "T, SBER, GAZP...",
	"target_price": "<целевая_цена_или_диапазон>",
	"target_change_percent": 0.0,
	"period": "Сегодня" | "Краткосрочный" | "Среднесрочный" | "Долгосрочный" | "Неопределенный",
	"full_text": "%s",
	"recommendation":"Покупать" | "Продавать " | "Держать" | "Неопределенный"
	}
	]

	Типы прогнозов (prediction_type)

	"Продолжение тренда" - Актив, вероятно, продолжит движение в текущем направлении (рост или падение).

	"Разворот" - Ожидается смена текущего тренда на противоположный.

	"Цель с коррекцией" - Актив достигнет определенной цели, но после этого возможна коррекция.

	"Накопление перед пробоем" - Актив находится в боковике, но готовится к резкому выходу из него (пробою).

	"Долгосрочный пессимизм" - Долгосрочные негативные ожидания по активу.

	"Неопределенный" - Прогноз не содержит четкого направления или типа.

	Временные периоды (period)

	"Сегодня" - Прогноз на текущий торговый день.
	"Краткосрочный" - От нескольких дней до нескольких недель.
	"Среднесрочный" - От нескольких недель до нескольких месяцев.
	"Долгосрочный" - От года и более.
	"Неопределенный" - Временной горизонт не указан.

	Тип рекомендации:
	"Покупать"
	"Продавать "
	"Держать"
	"Неопределенный" - когда не можешь дать рекомендацию

	Отвечай только JSON, без дополнительного текста.`, summary, messageID, summary)

	bodyBytes, err := c.sendOllamaRequest(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to send Ollama financial prediction request: %w", err)
	}

	jsonContent, err := c.parseOllamaResponse(bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Ollama financial prediction response: %w", err)
	}

	var financialPrediction FinancialPredictionPromptResponse
	if err := json.Unmarshal([]byte(jsonContent), &financialPrediction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal financial prediction JSON: %w, content: %s", err, jsonContent)
	}

	return &financialPrediction, nil
}

func (c *OllamaClient) AnalyzeMessage(ctx context.Context, message, channel string) (*MessageAnalysis, error) {
	// Шаг 1: Резюмирование сообщения
	summary, err := c.SummarizeMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to summarize message: %w", err)
	}

	// Генерируем уникальный ID для сообщения
	messageID := "some_unique_id" // TODO: Заменить на реальную генерацию ID

	// Шаг 2: Финансовый прогноз на основе резюме
	predictions, err := c.PredictFinancialOutcome(ctx, summary, channel, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to predict financial outcome: %w", err)
	}

	return &MessageAnalysis{
		Summary:     summary,
		Predictions: *predictions,
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

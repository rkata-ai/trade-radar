package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestPredictionStep_Execute проверяет функцию Execute для PredictionStep
func TestPredictionStep_Execute(t *testing.T) {
	// Внутренний JSON-массив, который будет в поле "response" ответа Ollama
	internalJSONResponse := `[
  {
    "message_id": "e7792424-c1d7-4141-a1883b6ad47f",
    "prediction_type": "Продолжение тренда",
    "ticker": "AFLT",
    "period": "Краткосрочный",
    "target_price": null,
    "target_change_percent": null,
    "recommendation": "Покупать",
    "direction": "Лонг",
    "justification_text": "Автор сообщения уверен в продолжении тренда,  оценивает силу продавцов и прогнозирует успех в декабрьских максимумах."
  },
  {
    "message_id": "e7792424-c1d7-4141-a1883b6ad47f",
    "prediction_type": "Разворот",
    "ticker": "AFLT",
    "period": "Среднесрочный",
    "target_price": null,
    "target_change_percent": null,
    "recommendation": "Неопределенный",
    "direction": "Неопределенный",
    "justification_text": "Автор сообщения указывает на необходимость пробития нисходящей трендовой линии."
  },
  {
    "message_id": "e7792424-c1d7-4141-a1883b6ad47f",
    "prediction_type": "Накопление перед пробоем",
    "ticker": "AFLT",
    "period": "Долгосрочный",
    "target_price": null,
    "target_change_percent": null,
    "recommendation": "Покупать",
    "direction": "Лонг",
    "justification_text": "Автор сообщения говорит о возможности обнуления депозита, что создает благоприятные условия для дальнейшего роста."
  },
  {
    "message_id": "e7792424-c1d7-4141-a1883b6ad47f",
    "prediction_type": "Разворот",
    "ticker": "AFLT",
    "period": "Долгосрочный",
    "target_price": null,
    "target_change_percent": null,
    "recommendation": "Неопределенный",
    "direction": "Неопределенный",
    "justification_text": "Автор сообщения подразумевает, что начнется нисходящий цикл"
  },
  {
    "message_id": "e7792424-c1d7-4141-a1883b6ad47f",
    "prediction_type": "Долгосрочный",
    "ticker": "AFLT",
    "period": "Неопределенный",
    "target_price": null,
    "target_change_percent": null,
    "recommendation": "Держать",
    "direction": "Лонг",
    "justification_text": "Автор сообщения предсказывает, что рынок продолжит расти."
  },
  {
    "message_id": "e7792424-c1d7-4141-a1883b6ad47f",
    "prediction_type": "Долгосрочный",
    "ticker": "SOL",
    "period": "Неопределенный",
    "target_price": null,
    "target_change_percent": null,
    "recommendation": "Держать",
    "direction": "Лонг",
    "justification_text": "Автор сообщения не указывает, что риск формирует ГиП, но считает, что у Эфириума есть шанс на рост"
  }
]`

	t.Run("Успешное выполнение и демаршалинг нескольких прогнозов", func(t *testing.T) {
		messageID := uuid.MustParse("2238f6c5-98a0-4180-8dcc-be76dd15fecc")

		formattedInternalResponse := internalJSONResponse // Теперь без Sprintf

		// Мок-функция для sendOllamaRequest
		mockSendOllamaRequest := func(ctx context.Context, prompt string) ([]byte, error) {
			fullOllamaResponse := OllamaGenerateResponse{
				Model:     "test-model",
				CreatedAt: "2023-11-20T17:28:43.078499Z",
				Response:  formattedInternalResponse,
				Done:      true,
			}
			return json.Marshal(fullOllamaResponse)
		}

		// Создаем клиент Ollama с мок-функцией
		client := &OllamaClient{
			sendRequestFunc: mockSendOllamaRequest,
			debug:           true,
		}

		// Создаем шаг прогнозирования
		step := NewPredictionStep()

		// Вызываем Execute
		predictions, err := step.Execute(context.Background(), client, "Тестовое сообщение", messageID)

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, predictions)
		assert.Len(t, predictions, 6)

		// Проверка первого прогноза
		assert.Equal(t, "Продолжение тренда", predictions[0].PredictionType)
		assert.Equal(t, "AFLT", predictions[0].Ticker)
		assert.True(t, predictions[0].TargetPrice.IsNull)
		assert.Equal(t, "", predictions[0].TargetPrice.String())
		assert.True(t, predictions[0].TargetChangePercent.IsNull)
		assert.Equal(t, "", predictions[0].TargetChangePercent.String())
		assert.Equal(t, "Покупать", predictions[0].Recommendation)
		assert.Equal(t, "Лонг", predictions[0].Direction)
		assert.Contains(t, predictions[0].JustificationText, "продолжении тренда")

		// Проверка второго прогноза
		assert.Equal(t, "Разворот", predictions[1].PredictionType)
		assert.Equal(t, "AFLT", predictions[1].Ticker)
		assert.True(t, predictions[1].TargetPrice.IsNull)
		assert.Equal(t, "", predictions[1].TargetPrice.String())
		assert.True(t, predictions[1].TargetChangePercent.IsNull)
		assert.Equal(t, "", predictions[1].TargetChangePercent.String())
		assert.Equal(t, "Неопределенный", predictions[1].Recommendation)
		assert.Equal(t, "Неопределенный", predictions[1].Direction)
		assert.Contains(t, predictions[1].JustificationText, "необходимость пробития нисходящей трендовой линии")

		// Проверка третьего прогноза
		assert.Equal(t, "Накопление перед пробоем", predictions[2].PredictionType)
		assert.Equal(t, "AFLT", predictions[2].Ticker)
		assert.True(t, predictions[2].TargetPrice.IsNull)
		assert.Equal(t, "", predictions[2].TargetPrice.String())
		assert.True(t, predictions[2].TargetChangePercent.IsNull)
		assert.Equal(t, "", predictions[2].TargetChangePercent.String())
		assert.Equal(t, "Покупать", predictions[2].Recommendation)
		assert.Equal(t, "Лонг", predictions[2].Direction)
		assert.Contains(t, predictions[2].JustificationText, "возможности обнуления депозита")

		// Проверка четвертого прогноза
		assert.Equal(t, "Разворот", predictions[3].PredictionType)
		assert.Equal(t, "AFLT", predictions[3].Ticker)
		assert.True(t, predictions[3].TargetPrice.IsNull)
		assert.Equal(t, "", predictions[3].TargetPrice.String())
		assert.True(t, predictions[3].TargetChangePercent.IsNull)
		assert.Equal(t, "", predictions[3].TargetChangePercent.String())
		assert.Equal(t, "Неопределенный", predictions[3].Recommendation)
		assert.Equal(t, "Неопределенный", predictions[3].Direction)
		assert.Contains(t, predictions[3].JustificationText, "начнется нисходящий цикл")

		// Проверка пятого прогноза
		assert.Equal(t, "Долгосрочный", predictions[4].PredictionType)
		assert.Equal(t, "AFLT", predictions[4].Ticker)
		assert.True(t, predictions[4].TargetPrice.IsNull)
		assert.Equal(t, "", predictions[4].TargetPrice.String())
		assert.True(t, predictions[4].TargetChangePercent.IsNull)
		assert.Equal(t, "", predictions[4].TargetChangePercent.String())
		assert.Equal(t, "Держать", predictions[4].Recommendation)
		assert.Equal(t, "Лонг", predictions[4].Direction)
		assert.Contains(t, predictions[4].JustificationText, "рынок продолжит расти")

		// Проверка шестого прогноза
		assert.Equal(t, "Долгосрочный", predictions[5].PredictionType)
		assert.Equal(t, "SOL", predictions[5].Ticker)
		assert.True(t, predictions[5].TargetPrice.IsNull)
		assert.Equal(t, "", predictions[5].TargetPrice.String())
		assert.True(t, predictions[5].TargetChangePercent.IsNull)
		assert.Equal(t, "", predictions[5].TargetChangePercent.String())
		assert.Equal(t, "Держать", predictions[5].Recommendation)
		assert.Equal(t, "Лонг", predictions[5].Direction)
		assert.Contains(t, predictions[5].JustificationText, "у Эфириума есть шанс на рост")
	})

	t.Run("Ошибка при запросе к Ollama", func(t *testing.T) {
		messageID := uuid.New()

		// Мок-функция, которая всегда возвращает ошибку
		mockSendOllamaRequest := func(ctx context.Context, prompt string) ([]byte, error) {
			return nil, fmt.Errorf("ошибка Ollama API")
		}

		// Создаем клиент Ollama с мок-функцией
		client := &OllamaClient{
			sendRequestFunc: mockSendOllamaRequest,
		}
		step := NewPredictionStep()

		predictions, err := step.Execute(context.Background(), client, "Тестовое сообщение", messageID)

		assert.Error(t, err)
		assert.Nil(t, predictions)
		assert.Contains(t, err.Error(), "failed to send ollama request")
	})

	t.Run("Некорректный JSON-ответ от Ollama", func(t *testing.T) {
		messageID := uuid.New()
		// Мок-функция
		mockSendOllamaRequest := func(ctx context.Context, prompt string) ([]byte, error) {
			fullOllamaResponse := OllamaGenerateResponse{
				Model:     "test-model",
				CreatedAt: "2023-11-20T17:28:43.078499Z",
				Response:  "{not a valid json}", // Некорректный JSON
				Done:      true,
			}
			return json.Marshal(fullOllamaResponse)
		}

		// Создаем клиент Ollama с мок-функцией
		client := &OllamaClient{
			sendRequestFunc: mockSendOllamaRequest,
		}
		step := NewPredictionStep()

		predictions, err := step.Execute(context.Background(), client, "Тестовое сообщение", messageID)

		assert.Error(t, err)
		assert.Nil(t, predictions)
		assert.Contains(t, err.Error(), "invalid JSON response from Ollama")
	})

	t.Run("Пустой JSON-контент от Ollama", func(t *testing.T) {
		messageID := uuid.New()
		// Мок-функция
		mockSendOllamaRequest := func(ctx context.Context, prompt string) ([]byte, error) {
			fullOllamaResponse := OllamaGenerateResponse{
				Model:     "test-model",
				CreatedAt: "2023-11-20T17:28:43.078499Z",
				Response:  "", // Пустой ответ
				Done:      true,
			}
			return json.Marshal(fullOllamaResponse)
		}

		// Создаем клиент Ollama с мок-функцией
		client := &OllamaClient{
			sendRequestFunc: mockSendOllamaRequest,
		}
		step := NewPredictionStep()

		predictions, err := step.Execute(context.Background(), client, "Тестовое сообщение", messageID)

		assert.Error(t, err)
		assert.Nil(t, predictions)
		assert.Contains(t, err.Error(), "ollama returned empty or null JSON content")
	})

	t.Run("Ollama возвращает пустой массив прогнозов", func(t *testing.T) {
		messageID := uuid.New()
		// Мок-функция
		mockSendOllamaRequest := func(ctx context.Context, prompt string) ([]byte, error) {
			fullOllamaResponse := OllamaGenerateResponse{
				Model:     "test-model",
				CreatedAt: "2023-11-20T17:28:43.078499Z",
				Response:  "[]", // Пустой массив JSON
				Done:      true,
			}
			return json.Marshal(fullOllamaResponse)
		}

		// Создаем клиент Ollama с мок-функцией
		client := &OllamaClient{
			sendRequestFunc: mockSendOllamaRequest,
		}
		step := NewPredictionStep()

		predictions, err := step.Execute(context.Background(), client, "Тестовое сообщение", messageID)

		assert.Error(t, err)
		assert.Nil(t, predictions)
		assert.Contains(t, err.Error(), "returned no predictions")
	})
}

# Система рейтинга телеграм-каналов по трейдингу

R&D исследование для создания системы оценки эффективности торговых рекомендаций из Telegram-каналов.

## 🎯 Цель проекта

Создать систему, которая:
1. Извлекает торговые рекомендации из Telegram-каналов
2. Получает исторические данные котировок через MetaTrader
3. Анализирует точность предсказаний
4. Рассчитывает рейтинг каналов

## 🏗️ Архитектура

```
traiding/
├── cmd/main.go              # Главный файл для R&D исследования
├── internal/
│   ├── config/              # Конфигурация
│   ├── telegram/            # Работа с Telegram API
│   ├── metatrader/          # Интеграция с MetaTrader
│   └── logger/              # Логирование
├── configs/                 # Конфигурационные файлы
└── go.mod                   # Зависимости Go
```

## 🚀 Быстрый старт

### Предварительные требования

1. **Go 1.21+** - [скачать](https://golang.org/dl/)
2. **Telegram Bot Token** - получить у [@BotFather](https://t.me/botfather)
3. **MetaTrader 5/4** - установленный терминал

### Установка зависимостей

```bash
go mod tidy
```

### Настройка конфигурации

1. Скопируйте и отредактируйте конфигурационный файл:
```bash
cp configs/config.yaml configs/config.local.yaml
```

2. Заполните необходимые параметры:
```yaml
telegram:
  bot_token: "YOUR_BOT_TOKEN_HERE"
  channels:
    - "your_trading_channel_1"
    - "your_trading_channel_2"

metatrader:
  server: "localhost"
  login: "YOUR_LOGIN"
  password: "YOUR_PASSWORD"
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

### Переменные окружения

```bash
# Telegram
export TRADING_TELEGRAM_BOT_TOKEN="your_bot_token"
export TRADING_TELEGRAM_API_ID="your_api_id"
export TRADING_TELEGRAM_API_HASH="your_api_hash"

# MetaTrader
export TRADING_METATRADER_SERVER="localhost"
export TRADING_METATRADER_LOGIN="your_login"
export TRADING_METATRADER_PASSWORD="your_password"

# Логирование
export LOG_LEVEL="debug"
```

### Структура конфигурации

```yaml
telegram:
  bot_token: "BOT_TOKEN"           # Токен бота от @BotFather
  api_id: "API_ID"                 # Telegram API ID
  api_hash: "API_HASH"             # Telegram API Hash
  channels:                         # Список каналов для анализа
    - "channel_username_1"
    - "channel_username_2"

metatrader:
  server: "localhost"               # Адрес сервера MT
  port: 443                         # Порт подключения
  login: "LOGIN"                    # Логин аккаунта
  password: "PASSWORD"              # Пароль аккаунта
  timeout: 30                       # Таймаут в секундах

logging:
  level: "info"                     # Уровень логирования
  format: "text"                    # Формат логов
  output: "stdout"                  # Вывод логов
```

## 📊 Что тестируется

### Telegram интеграция
- ✅ Подключение к Telegram Bot API
- ✅ Получение информации о каналах
- ✅ Извлечение сообщений из каналов
- ✅ Анализ структуры сообщений

### MetaTrader интеграция
- ✅ Подключение к терминалу
- ✅ Получение текущих котировок
- ✅ Загрузка исторических данных
- ✅ Работа с различными таймфреймами

## 🔍 Анализ данных

### Извлечение рекомендаций
- Поиск упоминаний акций/валют
- Определение типа рекомендации (buy/sell/hold)
- Извлечение временных рамок
- Анализ уверенности в рекомендации

### Расчет метрик
- Процент успешных предсказаний
- Средняя доходность
- Максимальная просадка
- Волатильность результатов

## 🚧 Текущий статус

- [x] Базовая структура проекта
- [x] Конфигурация и логирование
- [x] Telegram API клиент (базовый)
- [x] MetaTrader клиент (симуляция)
- [ ] Реальное подключение к MetaTrader
- [ ] Парсинг торговых рекомендаций
- [ ] Алгоритмы расчета рейтинга
- [ ] Веб-интерфейс

## 🐛 Известные проблемы

1. **Telegram API ограничения** - Bot API не может читать сообщения из каналов без подписки
2. **MetaTrader интеграция** - Требует дополнительных библиотек для реального подключения
3. **Парсинг рекомендаций** - Нужна интеграция с LLM для анализа текста

## 🔮 Следующие шаги

1. **Интеграция с MT5 API** - использование официального API MetaTrader 5
2. **Telegram Client API** - переход на MTProto для чтения каналов
3. **LLM интеграция** - подключение к Cursor API для анализа текста
4. **База данных** - сохранение и анализ исторических данных

## 📝 Лицензия

MIT License

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для новой функции
3. Внесите изменения
4. Создайте Pull Request

## 📞 Поддержка

По вопросам и предложениям создавайте Issues в репозитории.


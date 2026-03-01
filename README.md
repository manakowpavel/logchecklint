# logchecklint

Go-линтер для проверки лог-записей, совместимый с [golangci-lint](https://golangci-lint.run/).

Анализирует вызовы логгеров (`log/slog`, `go.uber.org/zap`) и проверяет сообщения на соответствие правилам.

## Правила

| # | Правило | Описание |
|---|---------|----------|
| 1 | Строчная буква | Лог-сообщение должно начинаться со строчной буквы |
| 2 | Только английский | Лог-сообщение должно быть только на английском языке |
| 3 | Без спецсимволов | Лог-сообщение не должно содержать спецсимволы или эмодзи |
| 4 | Без чувствительных данных | Лог-сообщение не должно содержать потенциально чувствительные данные |

### Примеры

```go
// Неправильно
slog.Info("Starting server on port 8080")  // начинается с заглавной буквы
slog.Info("запуск сервера")                // не на английском
slog.Info("server started!🚀")             // содержит спецсимволы и эмодзи
slog.Info("user password: " + password)    // содержит чувствительные данные

// Правильно
slog.Info("starting server on port 8080")
slog.Info("starting server")
slog.Info("server started")
slog.Info("user authenticated successfully")
```

## Установка и запуск

### Как standalone-линтер

```bash
# Сборка
go build -o bin/logchecklint ./cmd/logchecklint

# Запуск на проекте
./bin/logchecklint ./...
```

### Как плагин для golangci-lint (Module Plugin System)

**Рекомендуемый способ.** Используется [Module Plugin System](https://golangci-lint.run/docs/plugins/module-plugins/).

#### 1. Создайте файл `.custom-gcl.yml` в корне вашего проекта

```yaml
version: v2.10.1

plugins:
  - module: 'github.com/manakovpavel/logchecklint'
    import: 'github.com/manakovpavel/logchecklint'
    version: v0.1.0
```

#### 2. Создайте/обновите `.golangci.yml`

```yaml
version: "2"

linters:
  default: none
  enable:
    - logchecklint

  settings:
    custom:
      logchecklint:
        type: module
        description: "Linter for checking log messages"
        original-url: github.com/manakovpavel/logchecklint
        settings:
          disable_lowercase_check: false
          disable_english_check: false
          disable_special_char_check: false
          disable_sensitive_check: false
          custom_sensitive_keywords:
            - "bank_account"
            - "internal_id"
```

#### 3. Соберите и запустите

```bash
# Сборка кастомного golangci-lint с плагином
golangci-lint custom

# Запуск
./custom-gcl run ./...
```

## Конфигурация

Через секцию `settings` в `.golangci.yml` можно настроить:

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|--------------|----------|
| `disable_lowercase_check` | `bool` | `false` | Отключить проверку на строчную букву |
| `disable_english_check` | `bool` | `false` | Отключить проверку на английский язык |
| `disable_special_char_check` | `bool` | `false` | Отключить проверку на спецсимволы |
| `disable_sensitive_check` | `bool` | `false` | Отключить проверку на чувствительные данные |
| `custom_sensitive_keywords` | `[]string` | `[]` | Дополнительные ключевые слова для проверки |

## Поддерживаемые логгеры

- **log/slog** — `slog.Info()`, `slog.Error()`, `slog.Warn()`, `slog.Debug()`, `slog.Log()`
- **go.uber.org/zap** — `logger.Info()`, `logger.Error()`, `sugar.Infow()`, `sugar.Infof()` и другие
- **Любой логгер** с вызовами вида `log.Info("message")`, `logger.Error("message")`

## Авто-исправление (SuggestedFixes)

Линтер поддерживает автоматическое исправление для:
- **Правило 1**: Автоматически заменяет первую заглавную букву на строчную
- **Правило 3**: Автоматически удаляет спецсимволы и эмодзи

Для применения авто-исправлений используйте флаг `--fix`:

```bash
./custom-gcl run --fix ./...
```

## Разработка

```bash
# Клонирование
git clone https://github.com/manakovpavel/logchecklint.git
cd logchecklint

# Установка зависимостей
make deps

# Запуск тестов
make test

# Сборка standalone-бинарника
make build

# Сборка плагина для golangci-lint
make build-plugin
```

## Структура проекта

```
logchecklint/
├── cmd/logchecklint/         # standalone entrypoint
│   └── main.go
├── pkg/analyzer/             # основная логика анализатора
│   ├── analyzer.go           # анализатор (AST traversal, reporting)
│   ├── analyzer_test.go      # интеграционные тесты (analysistest)
│   ├── rules.go              # реализация правил проверки
│   └── rules_test.go         # unit-тесты правил
├── exampleapp/               # пример приложения с логами
├── plugin.go                 # регистрация плагина для golangci-lint
├── testdata/                 # тестовые данные для analysistest
├── .custom-gcl.yml           # конфигурация сборки плагина
├── .golangci.yml             # конфигурация golangci-lint
├── .github/workflows/ci.yml  # CI/CD пайплайн
├── Makefile                  # команды сборки и тестирования
├── go.mod
└── README.md
```

## Пример использования

Пример кода (фрагмент `exampleapp/main.go`):

```go
slog.Info("Starting server on port 8080")
slog.Info("запуск сервера")
slog.Info("server started!🚀")
slog.Info("user password: " + password)
slog.Info("starting server on port 8080")
slog.Info("server started")
slog.Info("user authenticated successfully")
```

Запуск линтера на exampleapp:

```bash
cd exampleapp
../bin/logchecklint ./...
```

Пример вывода:

```text
exampleapp/main.go:14:15: log message should start with a lowercase letter
exampleapp/main.go:15:15: log message should start with a lowercase letter
exampleapp/main.go:15:15: log message should be in English only
exampleapp/main.go:16:15: log message should be in English only
exampleapp/main.go:16:15: log message should not contain special characters or emoji
exampleapp/main.go:17:15: log message may contain sensitive data
exampleapp/main.go:28:20: log message should start with a lowercase letter
```

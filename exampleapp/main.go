package main

import (
    "log/slog"
    "os"
    "go.uber.org/zap"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    // примеры с нарушениями
    slog.Info("Starting server on port 8080")         // заглавная буква
    slog.Info("запуск сервера")                       // не английский
    slog.Warn("server started!🚀")                    // спецсимволы/эмодзи
    slog.Info("user password: secret123")             // чувствительные данные

    // корректные примеры
    slog.Info("starting server on port 8080")
    slog.Info("server started")
    slog.Info("user authenticated successfully")

    // Пример с zap
    zapLogger, _ := zap.NewProduction()
    defer func() {
        _ = zapLogger.Sync()
    }()

    zapLogger.Info("Starting zap logger")             // заглавная буква
    zapLogger.Info("starting zap logger ok")
}

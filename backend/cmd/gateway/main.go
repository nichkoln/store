// cmd/gateway/main.go
package main

import (
	"backend/internal/config"
	"backend/internal/gateway"
	"backend/pkg/logger"
	"fmt"
)

func main() {
	// Инициализация логирования
	logger.Init()

	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Создание API Gateway
	gw, err := gateway.NewGateway(cfg)
	if err != nil {
		logger.ErrorLogger.Fatal("Ошибка инициализации Gateway:", err)
	}

	// Настройка маршрутов
	router := gw.SetupRouter()

	// Запуск API Gateway
	if err := router.Run(fmt.Sprintf(":%s", cfg.APIProxyPort)); err != nil {
		logger.ErrorLogger.Fatal("Ошибка запуска API Gateway:", err)
	}
}

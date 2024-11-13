// internal/gateway/handlers/order_handlers.go
package handlers

import (
	"backend/internal/config"
	"backend/pkg/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	StrapiURL string
}

func NewOrderHandler(cfg *config.Config) *OrderHandler {
	return &OrderHandler{
		StrapiURL: cfg.StrapiURL,
	}
}

// CreateOrder godoc
// @Summary Создать новый заказ
// @Description Создаёт новый заказ для текущего пользователя
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body map[string]interface{} true "Данные заказа"
// @Success 201 {object} interface{}
// @Failure 400 {object} gin.H{"error": "Неверные данные"}
// @Failure 401 {object} gin.H{"error": "Не авторизован"}
// @Failure 500 {object} gin.H{"error": "Ошибка сервера"}
// @Router /api/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	var orderData map[string]interface{}
	if err := c.ShouldBindJSON(&orderData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	// Добавляем ID пользователя к данным заказа
	orderData["user"] = userID

	payload := map[string]interface{}{
		"data": orderData,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.ErrorLogger.Println("Ошибка маршалинга данных заказа:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/orders", h.StrapiURL), "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.ErrorLogger.Println("Ошибка отправки запроса к Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		logger.ErrorLogger.Printf("Strapi ответил с ошибкой %d: %s", resp.StatusCode, string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorLogger.Println("Ошибка чтения ответа от Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	var createdOrder interface{}
	if err := json.Unmarshal(body, &createdOrder); err != nil {
		logger.ErrorLogger.Println("Ошибка парсинга ответа Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	c.JSON(http.StatusCreated, createdOrder)
}

// GetOrders godoc
// @Summary Получить список заказов
// @Description Возвращает список всех заказов текущего пользователя
// @Tags Orders
// @Accept json
// @Produce json
// @Success 200 {object} interface{}
// @Failure 401 {object} gin.H{"error": "Не авторизован"}
// @Failure 500 {object} gin.H{"error": "Ошибка сервера"}
// @Router /api/orders [get]
func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/orders?filters[user][$eq]=%d&populate=*", h.StrapiURL, userID))
	if err != nil {
		logger.ErrorLogger.Println("Ошибка запроса к Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logger.ErrorLogger.Printf("Strapi ответил с ошибкой %d: %s", resp.StatusCode, string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorLogger.Println("Ошибка чтения ответа от Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	var orders interface{}
	if err := json.Unmarshal(body, &orders); err != nil {
		logger.ErrorLogger.Println("Ошибка парсинга ответа Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

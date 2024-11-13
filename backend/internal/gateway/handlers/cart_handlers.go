// internal/gateway/handlers/cart_handlers.go
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

type CartHandler struct {
	StrapiURL string
}

func NewCartHandler(cfg *config.Config) *CartHandler {
	return &CartHandler{
		StrapiURL: cfg.StrapiURL,
	}
}

// Получение содержимого корзины
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/carts?filters[user][$eq]=%d&populate=*", h.StrapiURL, userID))
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

	var cart interface{}
	if err := json.Unmarshal(body, &cart); err != nil {
		logger.ErrorLogger.Println("Ошибка парсинга ответа Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// Добавление товара в корзину
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	var addData struct {
		ProductID int `json:"productId"`
		Quantity  int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&addData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"user":     userID,
			"product":  addData.ProductID,
			"quantity": addData.Quantity,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.ErrorLogger.Println("Ошибка маршалинга данных корзины:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/carts", h.StrapiURL), "application/json", bytes.NewBuffer(payloadBytes))
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

	var addedCart interface{}
	if err := json.Unmarshal(body, &addedCart); err != nil {
		logger.ErrorLogger.Println("Ошибка парсинга ответа Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	c.JSON(http.StatusCreated, addedCart)
}

// Удаление товара из корзины
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	var removeData struct {
		CartItemID int `json:"cartItemId"`
	}

	if err := c.ShouldBindJSON(&removeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/carts/%d", h.StrapiURL, removeData.CartItemID), nil)
	if err != nil {
		logger.ErrorLogger.Println("Ошибка создания запроса к Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.ErrorLogger.Println("Ошибка отправки запроса к Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		logger.ErrorLogger.Printf("Strapi ответил с ошибкой %d: %s", resp.StatusCode, string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Товар удалён из корзины"})
}

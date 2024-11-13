// internal/gateway/handlers/auth_handlers.go
package handlers

import (
	"backend/internal/config"
	"backend/pkg/logger"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	StrapiURL string
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		StrapiURL: cfg.StrapiURL,
	}
}

// GetCurrentUser godoc
// @Summary Получить информацию о текущем пользователе
// @Description Возвращает данные текущего пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} interface{}
// @Failure 401 {object} gin.H{"error": "Не авторизован"}
// @Failure 500 {object} gin.H{"error": "Ошибка сервера"}
// @Router /api/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	// Предполагается, что userID это числовой ID пользователя
	uid, err := strconv.Atoi(fmt.Sprintf("%v", userID))
	if err != nil {
		logger.ErrorLogger.Println("Неверный формат userID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/users/%d", h.StrapiURL, uid))
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

	var user interface{}
	if err := json.Unmarshal(body, &user); err != nil {
		logger.ErrorLogger.Println("Ошибка парсинга ответа Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Дополнительные хендлеры для регистрации и входа можно добавить здесь

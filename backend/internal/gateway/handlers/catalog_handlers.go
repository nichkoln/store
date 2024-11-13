// internal/gateway/handlers/catalog_handlers.go
package handlers

import (
	"backend/internal/config"
	"backend/pkg/logger"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CatalogHandler struct {
	StrapiURL string
}

type Product struct {
	ID          int         `json:"id"`
	DocumentID  string      `json:"documentId"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       int         `json:"price"`
	Category    string      `json:"category"`
	Brand       string      `json:"brand"`
	Size        interface{} `json:"size"`
	CreatedAt   string      `json:"createdAt"`
	UpdatedAt   string      `json:"updatedAt"`
	PublishedAt string      `json:"publishedAt"`
	ImageURL    string      `json:"imageUrl"` // Добавляем поле для URL изображения
	Images      []Image     `json:"image"`    // Полный массив изображений
}

type Image struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	URL     string  `json:"url"`
	Formats Formats `json:"formats"`
}

type Formats struct {
	Thumbnail Thumbnail `json:"thumbnail"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

func NewCatalogHandler(cfg *config.Config) *CatalogHandler {
	return &CatalogHandler{
		StrapiURL: cfg.StrapiURL,
	}
}

// Получение списка товаров
func (h *CatalogHandler) GetProducts(c *gin.Context) {
	resp, err := http.Get(fmt.Sprintf("%s/api/products?populate=image", h.StrapiURL))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorLogger.Println("Ошибка чтения ответа от Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	var rawResponse struct {
		Data []Product `json:"data"`
	}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		logger.ErrorLogger.Println("Ошибка парсинга ответа Strapi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	for i := range rawResponse.Data {
		product := &rawResponse.Data[i]
		if len(product.Images) > 0 {
			product.ImageURL = product.Images[0].URL // URL первого изображения
		}
		switch v := product.Size.(type) {
		case string:
			product.Size = []string{v} // Одиночный размер
		case []interface{}:
			var sizes []string
			for _, size := range v {
				if sizeStr, ok := size.(string); ok {
					sizes = append(sizes, sizeStr)
				}
			}
			product.Size = sizes
		}
	}

	c.JSON(http.StatusOK, rawResponse)
}

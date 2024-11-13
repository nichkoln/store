// internal/gateway/gateway.go
package gateway

import (
	"backend/internal/config"
	"backend/internal/gateway/handlers"
	"backend/pkg/logger"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/ulule/limiter/v3"
	ginmiddleware "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
)

type Gateway struct {
	StrapiURL   *url.URL
	BudibaseURL *url.URL
	JWTSecret   []byte

	AuthHandler    *handlers.AuthHandler
	CatalogHandler *handlers.CatalogHandler
	CartHandler    *handlers.CartHandler
	OrderHandler   *handlers.OrderHandler
}

// NewGateway инициализирует новый API Gateway
func NewGateway(cfg *config.Config) (*Gateway, error) {
	strapiURL, err := url.Parse(cfg.StrapiURL)
	if err != nil {
		return nil, err
	}

	budibaseURL, err := url.Parse("http://budibase:80") // Адрес Budibase внутри Docker сети
	if err != nil {
		return nil, err
	}

	gw := &Gateway{
		StrapiURL:   strapiURL,
		BudibaseURL: budibaseURL,
		JWTSecret:   []byte(cfg.JWTSecret),

		AuthHandler:    handlers.NewAuthHandler(cfg),
		CatalogHandler: handlers.NewCatalogHandler(cfg),
		CartHandler:    handlers.NewCartHandler(cfg),
		OrderHandler:   handlers.NewOrderHandler(cfg),
	}

	return gw, nil
}

// Middleware проверяет JWT токен
func (g *Gateway) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		//parts := gin.H{}
		partsSlice := splitN(authHeader, " ", 2)
		if len(partsSlice) != 2 || partsSlice[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		tokenString := partsSlice[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверка метода подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return g.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Извлечение информации из токена
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("userID", claims["id"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		c.Next()
	}
}

// splitN разделяет строку по разделителю и возвращает массив строк
func splitN(s, sep string, n int) []string {
	if n <= 0 {
		return []string{}
	}
	return split(s, sep, n)
}

// SetupRouter настраивает маршруты API Gateway
func (g *Gateway) SetupRouter() *gin.Engine {
	router := gin.Default()

	// Настройка ограничения скорости
	rate, err := limiter.NewRateFromFormatted("100-M") // 100 запросов в минуту
	if err != nil {
		logger.ErrorLogger.Fatal("Ошибка создания лимитера:", err)
	}

	store := memory.NewStore()
	rateLimiter := ginmiddleware.NewMiddleware(limiter.New(store, rate))

	router.Use(rateLimiter)

	// Применение middleware для аутентификации
	router.Use(g.Middleware())

	// Регистрация маршрутов для авторизации
	authRoutes := router.Group("/api/auth")
	{
		authRoutes.GET("/me", g.AuthHandler.GetCurrentUser)
		// Добавьте другие маршруты авторизации
	}

	// Регистрация маршрутов для каталога
	catalogRoutes := router.Group("/api/catalog")
	{
		catalogRoutes.GET("/products", g.CatalogHandler.GetProducts)
		// Добавьте другие маршруты каталога
	}

	// Регистрация маршрутов для корзины
	cartRoutes := router.Group("/api/cart")
	{
		cartRoutes.GET("/", g.CartHandler.GetCart)
		cartRoutes.POST("/add", g.CartHandler.AddToCart)
		cartRoutes.POST("/remove", g.CartHandler.RemoveFromCart)
		// Добавьте другие маршруты корзины
	}

	// Регистрация маршрутов для заказов
	orderRoutes := router.Group("/api/orders")
	{
		orderRoutes.POST("/", g.OrderHandler.CreateOrder)
		orderRoutes.GET("/", g.OrderHandler.GetOrders)
		// Добавьте другие маршруты заказов
	}

	// Регистрация Swagger (если используется)
	//router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

package fbauth

import (
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/mousybusiness/googlecloudgo/pkg/secrets"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	authorizationHeader = "Authorization"
	apiKeyHeader        = "X-API-Key"
	cronExecutedHeader  = "X-Appengine-Cron"
)

// Gin middleware for JWT auth
func AuthJWT(client *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		authHeader := c.Request.Header.Get(authorizationHeader)
		log.Println("----->", authHeader)
		token := strings.Replace(authHeader, "Bearer ", "", 1)
		idToken, err := client.VerifyIDToken(c, token) // usually hits a local cache

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": http.StatusText(http.StatusUnauthorized),
			})

			return
		}

		log.Println(">>>AUTH TIME>>>", time.Since(startTime))

		c.Set(valName, idToken)
		c.Next()
	}
}

// Gin middleware for API key augth
func APIKeyAuth(secretId string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Request.Header.Get(apiKeyHeader)

		secret, err := secrets.GetSecret(secretId)
		if err != nil {
			log.Println("error while getting secret")
			_ = c.Error(status.Error(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))
			return
		}

		if secret != key {
			log.Println("key doesnt match!")
			_ = c.Error(status.Error(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))
			return
		}

		c.Next()
	}
}

// AppEngine cron authentication
func AuthAppEngineCron() gin.HandlerFunc {
	return func(c *gin.Context) {
		cron := c.Request.Header.Get(cronExecutedHeader)

		if cron != "true" {
			log.Println("Not called from cron - do not allow access")
			_ = c.Error(status.Error(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))
			return
		}

		c.Next()
	}
}

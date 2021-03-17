package auth

import (
	fbauth "firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/mousybusiness/googlecloudgo/pkg/secrets"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	authorizationHeader = "Authorization"
	apiKeyHeader        = "X-API-Key"
	cronExecutedHeader  = "X-Appengine-Cron"
)

// Gin middleware for JWT auth
func AuthJWT(client *fbauth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		authHeader := c.Request.Header.Get(authorizationHeader)
		token := strings.Replace(authHeader, "Bearer ", "", 1)
		idToken, err := client.VerifyIDToken(c, token) // usually hits a local cache
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		log.Println("Auth time:", time.Since(startTime))

		c.Set(FirebaseContextVal, idToken)
		c.Next()
	}
}

// API key auth middleware
func AuthAPIKey(secretId string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Request.Header.Get(apiKeyHeader)

		secret, err := secrets.GetSecret(secretId)
		if err != nil {
			log.Println("failed to get secret")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		if secret != key {
			log.Println("key doesnt match!")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		log.Println("no error during check")
		c.Next()
	}
}

// AppEngine cron authentication
func AuthAppEngineCron() gin.HandlerFunc {
	return func(c *gin.Context) {
		cron := c.Request.Header.Get(cronExecutedHeader)

		if cron != "true" {
			log.Println("not invoked by cron - access denied")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		c.Next()
	}
}

func checkInternal(ip string) bool {
	if strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.168") {
		return true
	}

	if  strings.HasPrefix(ip, "172.") {
		exp := regexp.MustCompile(`\d{1,3}\.(\d{1,3})\.\d{1,3}\.\d{1,3}`)
		v := exp.ReplaceAllString(ip, `$1`)
		atoi, err := strconv.Atoi(v)
		if err != nil {
			log.Println("failed to convert to int", v)
			return false
		}

		if atoi >= 16 && atoi <= 31 {
			return true
		}
	}

	return false
}

// only allow internal ip ranges
func AuthInternalOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ! checkInternal(ip) {
			log.Println("rejected IP:", ip)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "tisk tisk tisk",
			})
			return
		}

		c.Next()
	}
}
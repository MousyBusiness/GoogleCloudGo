package fbauth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

const (
	valName = "FIREBASE_ID_TOKEN"
)

var (
	firebaseConfigFile = os.Getenv("FIREBASE_CONFIG_FILE")
)

// load firebase configuration and create the auth client
// expecting env variable FIREBASE_CONFIG_FILE
func InitAuth() (*auth.Client, error) {
	opt := option.WithCredentialsFile(firebaseConfigFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app, creating fb app: %v", err)
	}

	client, errAuth := app.Auth(context.Background())
	if errAuth != nil {
		return nil, fmt.Errorf("error initializing auth, creating fb client: %v", errAuth)
	}
	return client, nil
}

// use id_token provided in Authorization: Bearer [ID_TOKEN]
func VerifyToken(c *gin.Context, client *auth.Client) (int, string, error) {
	authHeader := c.Request.Header.Get(authorizationHeader)
	token := strings.Replace(authHeader, "Bearer ", "", 1)

	idtoken, err := client.VerifyIDToken(c, token)
	if err != nil {
		log.Println("failed to verify token", err)
		return http.StatusUnauthorized, "not authenticated", err
	}

	claims := idtoken.Claims
	if admin, ok := claims["admin"]; ok {
		if admin.(bool) {
			return 0, "", nil
		}
	}

	log.Println("not admin")
	return http.StatusForbidden, "not admin", errors.New("not admin")
}

func ExtractClaims(c *gin.Context) (*auth.Token, error) {
	idToken, ok := c.Get(valName)
	if !ok {
		return nil, errors.New("Failed to extract claims")
	}
	return idToken.(*auth.Token), nil
}

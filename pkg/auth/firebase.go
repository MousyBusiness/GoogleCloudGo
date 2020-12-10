package auth

import (
	"context"
	"errors"
	errs "github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	fbauth "firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
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
func InitAuth() (*fbauth.Client, error) {
	if firebaseConfigFile == "" {
		return nil, errors.New("FIREBASE_CONFIG_FILE required")
	}

	opt := option.WithCredentialsFile(firebaseConfigFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, errs.Wrap(err, "error initializing app, creating fb app")
	}

	client, errAuth := app.Auth(context.Background())
	if errAuth != nil {
		return nil, errs.Wrap(err, "error initializing auth, creating fb client")
	}
	return client, nil
}

// use id_token provided in Authorization: Bearer [ID_TOKEN]
func VerifyToken(c *gin.Context, client *fbauth.Client) (int, string, error) {
	authHeader := c.Request.Header.Get(authorizationHeader)
	token := strings.Replace(authHeader, "Bearer ", "", 1)

	idtoken, err := client.VerifyIDToken(c, token)
	if err != nil {
		return http.StatusUnauthorized, "not authenticated", errs.Wrap(err, "failed to verify token")
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

func ExtractClaims(c *gin.Context) (*fbauth.Token, error) {
	idToken, ok := c.Get(valName)
	if !ok {
		return nil, errors.New("Failed to extract claims")
	}
	return idToken.(*fbauth.Token), nil
}

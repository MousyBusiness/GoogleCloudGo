package fbauth

import (
	"context"
	"errors"
	"firebase.google.com/go/auth"
	"log"
)

// add 'admin' claim to users JWT so they can perform admin actions (they still need to be authenticated with JWT)
func ElevateToAdmin(ctx context.Context, client *auth.Client, uid string) error {
	claims := map[string]interface{}{"admin": true}
	err := client.SetCustomUserClaims(ctx, uid, claims)
	if err != nil {
		log.Println("error setting custom claims %v\n", err)
		return err
	}
	return nil
}

// remove admin claim from users JWT so they can no longer perform admin actions
func RevokeAdmin(ctx context.Context, client *auth.Client, uid string) error {
	// setting claims to nil will remove admin rights
	err := client.SetCustomUserClaims(ctx, uid, nil)
	if err != nil {
		log.Println("error revoking custom claims %v\n", err)
		return err
	}
	return nil
}

// use user id to very if user is admin or not
func Verify(ctx context.Context, client *auth.Client, uid string) error {
	// get the user
	user, err := client.GetUser(ctx, uid)
	if err != nil {
		return err
	}

	// The claims can be accessed on the user record.
	if admin, ok := user.CustomClaims["admin"]; ok {
		if admin.(bool) {
			return nil
		}
	}
	return errors.New("not admin")
}

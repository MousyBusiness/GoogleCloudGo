package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	refreshURL = "https://securetoken.googleapis.com/v1/token"
)

type RefreshResponse struct {
	ExpiresIn    string `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
}

type RefreshToken string
type APIKey string

//https://firebase.google.com/docs/reference/rest/auth
func RefreshFirebaseToken(token RefreshToken, secret APIKey) (RefreshResponse, error) {

	b, err := json.Marshal(struct {
		RefreshToken string `json:"refresh_token"`
		GrantType    string `json:"grant_type"`
	}{
		RefreshToken: string(token),
		GrantType:    "refresh_token",
	})
	if err != nil {
		return RefreshResponse{}, err
	}

	resp, err := http.Post(fmt.Sprintf("%s?key=%s", refreshURL, secret), http.DetectContentType(b), bytes.NewReader(b))
	if err != nil {
		return RefreshResponse{}, err
	}

	code := resp.StatusCode
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RefreshResponse{}, err
	}

	if code != 200 {
		return RefreshResponse{}, errors.New(fmt.Sprintf("status code not 200, code: %d, error: %v", code, string(body)))
	}

	var r RefreshResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return RefreshResponse{}, err
	}

	return r, nil
}

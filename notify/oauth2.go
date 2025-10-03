package notify

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

var oauth2CodeCh = make(chan string)
var oauth2RediretServer *http.Server

func GetNotifyOAuth2TokenStep1() (string, error) {
	config := getNotifyOAuth2Config()
	if config == nil {
		return "", fmt.Errorf("no oauth2 config")
	}
	state := randCryptoString(32)
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("state") != state {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "OAuth2 authorization done. close this Window")
		oauth2CodeCh <- code
	})
	go func() {
		oauth2RediretServer = &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", datastore.NotifyOAuth2RedirectPort)}
		if err := oauth2RediretServer.ListenAndServe(); err != nil {
			log.Printf("oauth2RediretServer err=%v", err)
		}
	}()
	return url, nil
}

func GetNotifyOAuth2TokenStep2() error {
	defer func() {
		if oauth2RediretServer != nil {
			oauth2RediretServer.Close()
			oauth2RediretServer = nil
		}
	}()
	select {
	case code := <-oauth2CodeCh:
		config := getNotifyOAuth2Config()
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			return fmt.Errorf("fail to get token: %w", err)
		}
		datastore.SaveNotifyOAuth2Token(token)
	case <-time.NewTimer(time.Minute).C:
		return fmt.Errorf("get auth code timeout")
	}
	return nil
}

func getNotifyOAuth2Config() *oauth2.Config {
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", datastore.NotifyOAuth2RedirectPort)
	switch datastore.NotifyConf.Provider {
	case "google":
		return &oauth2.Config{
			ClientID:     datastore.NotifyConf.ClientID,
			ClientSecret: datastore.NotifyConf.ClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURL,
			Scopes:       []string{"https://mail.google.com/"},
		}
	case "microsoft":
		return &oauth2.Config{
			ClientID:     datastore.NotifyConf.ClientID,
			ClientSecret: datastore.NotifyConf.ClientSecret,
			Endpoint:     microsoft.AzureADEndpoint(datastore.NotifyConf.MSTenant),
			RedirectURL:  redirectURL,
			Scopes:       []string{"https://outlook.office.com/SMTP.Send"},
		}
	default:
		return nil
	}
}

func getNotifyOAuth2Token() *oauth2.Token {
	oldToken := datastore.GetNotifyOAuth2Token()
	if oldToken == nil {
		return nil
	}
	if oldToken.Valid() {
		return oldToken
	}
	config := getNotifyOAuth2Config()
	if config == nil {
		return nil
	}
	tokenSource := config.TokenSource(context.Background(), oldToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Printf("Fail to refresh token err=%v", err)
		return nil
	}
	log.Printf("oauth2 token updated old=%v new=%v", oldToken.Expiry, newToken.Expiry)
	datastore.SaveNotifyOAuth2Token(newToken)
	return newToken
}

func randCryptoString(length int) string {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "randamu_twsnmp"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

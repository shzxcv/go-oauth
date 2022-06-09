package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	code := authRequest()
	token := tokenRequest(code)
	fmt.Printf("Your token is %s", token.AccessToken)
}

var (
	clientID     string
	clientSecret string
	// random
	state    string = "xyz"
	authCode string
)

const (
	redirectURI   string = "http://127.0.0.1:6749/callback"
	authEndpoint  string = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenEndPoint string = "https://www.googleapis.com/oauth2/v4/token"
	scope         string = "https://www.googleapis.com/auth/photoslibrary.readonly"
)

func inputClientInfo() {
	fmt.Println("ClientID: ")
	fmt.Scan(&clientID)
	fmt.Println("ClientSecret: ")
	fmt.Scan(&clientSecret)
}

func timeoutHTTPServer(waitTime time.Duration) {
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)
	srv := startHTTPServer(httpServerExitDone)
	time.Sleep(waitTime * time.Second)
	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err)
	}
	httpServerExitDone.Wait()
}

func startHTTPServer(wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{Addr: ":6749"}
	http.HandleFunc("/callback", callback)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	return srv
}

func callback(w http.ResponseWriter, r *http.Request) {
	authCode = r.FormValue("code")
	if authCode != "" {
		w.Write([]byte(`You can get the auth code.
Close this tab and back the terminal.`))
	} else {
		w.Write([]byte(`You can't get the auth code.
Close this tab and check the code.`))
	}
}

func authRequest() string {
	inputClientInfo()

	authRequstURL := fmt.Sprintf("%s?response_type=code&client_id=%s&state=%s&scope=%s&redirect_uri=%s", authEndpoint, clientID, state, scope, redirectURI)
	err := exec.Command("open", authRequstURL).Start()
	if err != nil {
		log.Fatal(err)
	}

	timeoutHTTPServer(10)
	return authCode
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func tokenRequest(code string) *tokenResponse {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Add("client_secret", clientSecret)
	values.Add("redirect_uri", redirectURI)
	values.Add("grant_type", "authorization_code")
	values.Add("code", code)

	req, err := http.NewRequest("POST", tokenEndPoint, strings.NewReader(values.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	token := &tokenResponse{}
	if err := json.Unmarshal(body, &token); err != nil {
		log.Fatal(err)
	}
	return token
}

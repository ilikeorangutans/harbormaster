package azkaban

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrInvalidSessionID = fmt.Errorf("invalid session id, session might have expired")

func ConnectWithSessionID(url string, sessionID string) (*Client, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return &Client{
		http:      client,
		SessionID: sessionID,
		url:       url,
	}, nil
}

func ConnectWithUsernameAndPassword(url string, username string, password string) (*Client, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	q := req.URL.Query()
	q.Add("action", "login")
	q.Add("username", username)
	q.Add("password", password)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var status LoginResponse
	err = decoder.Decode(&status)
	if err != nil {
		return nil, err
	}

	if status.Status != "success" {
		return nil, fmt.Errorf("login failed")
	}

	return &Client{
		http:      client,
		SessionID: status.SessionID,
		url:       url,
	}, nil
}

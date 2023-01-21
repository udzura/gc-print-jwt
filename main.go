package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type accessToken struct {
	AccessToken string `json:"access_token"`
}

type token struct {
	Token string `json:"token"`
}

func main() {
	client := new(http.Client)

	if len(os.Args) != 2 {
		fmt.Println("Usage: gc-print-jwt FUNCTION_NAME")
		return
	}
	functionName := os.Args[1]

	req, _ := http.NewRequest(
		"GET",
		"http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/email",
		nil,
	)
	req.Header.Set("Metadata-Flavor", "Google")
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	email := string(data)

	req, _ = http.NewRequest(
		"GET",
		"http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token",
		nil,
	)
	req.Header.Set("Metadata-Flavor", "Google")
	res, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	t := accessToken{}
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		panic(err)
	}

	generateIdTokenUrl := fmt.Sprintf(
		"https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateIdToken",
		email,
	)
	jsonBody := fmt.Sprintf(
		`{"audience": "%s"}`,
		functionName,
	)
	reqBody := bytes.NewBuffer([]byte(jsonBody))
	req, _ = http.NewRequest(
		"POST",
		generateIdTokenUrl,
		reqBody,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))
	res, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	t2 := token{}
	if err := json.NewDecoder(res.Body).Decode(&t2); err != nil {
		panic(err)
	}

	fmt.Println(t2.Token)
}

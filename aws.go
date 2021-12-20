package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type profileResult struct {
	Result []profile `json:"result"`
}

type profile struct {
	Name string `json:"name"`
}

type accountResult struct {
	Result []account `json:"result"`
}

type account struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type credentials struct {
	RoleCredentials roleCredentials `json:"roleCredentials"`
}

type unhappyResponse struct {
	Message string `json:"message"`
	Type    string `json:"__type"`
}

type roleCredentials struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      int64  `json:"expiration"`
}

func getAWSResource(url, token string) ([]byte, error) {
	var bs []byte
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return bs, fmt.Errorf("got error %s", err.Error())
	}
	req.Header.Set("x-amz-sso-bearer-token", token)
	req.Header.Add("x-amz-sso_bearer_token", token)
	resp, err := client.Do(req)
	if err != nil {
		return bs, fmt.Errorf("got error %s", err.Error())
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getAccounts(url, token string) (accountResult, error) {
	bs, err := getAWSResource(url, token)
	if err != nil {
		log.Fatalln(err)
	}
	var accounts accountResult
	err = json.Unmarshal(bs, &accounts)
	if err != nil {
		return accounts, err
	}
	return accounts, nil
}

func getProfiles(url, token string) (profileResult, error) {
	bs, err := getAWSResource(url, token)
	if err != nil {
		log.Fatalln(err)
	}
	var profiles profileResult
	err = json.Unmarshal(bs, &profiles)
	if err != nil {
		return profiles, err
	}
	return profiles, nil
}

func getCredentials(url, token string) (credentials, error) {
	var c credentials
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c, fmt.Errorf("got error %s", err.Error())
	}
	req.Header.Set("x-amz-sso-bearer-token", token)
	req.Header.Add("x-amz-sso_bearer_token", token)
	resp, err := client.Do(req)
	if err != nil {
		return c, fmt.Errorf("got error %s", err.Error())
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c, fmt.Errorf("got error %s", err.Error())
	}
	var failureResponse unhappyResponse
	if err = json.Unmarshal(bs, &failureResponse); err != nil {
		log.Fatalln(err)
	}
	if strings.Contains(failureResponse.Type, "Forbidden") {
		return c, errors.New("invalid account name or profile name")
	} else if len(failureResponse.Message) > 0 {
		return c, fmt.Errorf("failed to get access credentials due to: %s", failureResponse.Message)
	}

	if err = json.Unmarshal(bs, &c); err != nil {
		log.Fatalln(err)
	}

	return c, nil
}

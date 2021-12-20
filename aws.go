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

type account struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type accounts struct {
	Result []account `json:"result"`
}

type profile struct {
	Name string `json:"name"`
}

type profiles struct {
	Result []profile `json:"result"`
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

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient
)

func init() {
	Client = &http.Client{
		Timeout: time.Second * 20,
	}
}

func getAWSResource(url, token string) ([]byte, error) {
	var bs []byte
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return bs, fmt.Errorf("got error %s", err.Error())
	}
	req.Header.Set("x-amz-sso-bearer-token", token)
	req.Header.Add("x-amz-sso_bearer_token", token)
	resp, err := Client.Do(req)
	if err != nil {
		return bs, fmt.Errorf("got error %s", err.Error())
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getAccounts(url, token string) (accounts, error) {
	bs, err := getAWSResource(url, token)
	if err != nil {
		log.Fatalln(err)
	}
	var accounts accounts
	err = json.Unmarshal(bs, &accounts)
	if err != nil {
		return accounts, err
	}
	return accounts, nil
}

func getProfiles(url, token string) (profiles, error) {
	bs, err := getAWSResource(url, token)
	if err != nil {
		log.Fatalln(err)
	}
	var profiles profiles
	err = json.Unmarshal(bs, &profiles)
	if err != nil {
		return profiles, err
	}
	return profiles, nil
}

func getCredentials(url, token string) (credentials, error) {
	var c credentials
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c, fmt.Errorf("got error %s", err.Error())
	}
	req.Header.Set("x-amz-sso-bearer-token", token)
	req.Header.Add("x-amz-sso_bearer_token", token)
	resp, err := Client.Do(req)
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

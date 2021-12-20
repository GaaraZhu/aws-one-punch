package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

type failureResponse struct {
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

type AWSService struct {
	client HTTPClient
}

func NewAWSService(client HTTPClient) AWSService {
	return AWSService{
		client: client,
	}
}

func (as *AWSService) getAccounts(url, token string) (accounts, error) {
	bs, err := as.getAWSResource(url, token)
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

func (as *AWSService) getProfiles(url, token string) (profiles, error) {
	bs, err := as.getAWSResource(url, token)
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

func (as *AWSService) getCredentials(url, token string) (credentials, error) {
	bs, err := as.getAWSResource(url, token)
	if err != nil {
		log.Fatalln(err)
	}
	var c credentials
	if err = json.Unmarshal(bs, &c); err != nil {
		log.Fatalln(err)
	}

	return c, nil
}

func (as *AWSService) getAWSResource(url, token string) ([]byte, error) {
	var bs []byte
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return bs, fmt.Errorf("got error %s", err.Error())
	}
	req.Header.Set("x-amz-sso-bearer-token", token)
	req.Header.Add("x-amz-sso_bearer_token", token)
	resp, err := as.client.Do(req)
	if err != nil {
		return bs, fmt.Errorf("got error %s", err.Error())
	}
	defer resp.Body.Close()
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return bs, err
	}

	// failure checking
	var failureResp failureResponse
	if err = json.Unmarshal(bs, &resp); err != nil {
		log.Fatalln(err)
	}
	if strings.Contains(failureResp.Type, "Forbidden") {
		return bs, errors.New("invalid account name or profile name")
	} else if len(failureResp.Message) > 0 {
		return bs, fmt.Errorf("operation failed due to: %s", failureResp.Message)
	}

	return bs, err
}

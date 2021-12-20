package main

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type HTTPClientMock struct {
	DoFunc func(*http.Request) (*http.Response, error)
}

func (H *HTTPClientMock) Do(r *http.Request) (*http.Response, error) {
	return H.DoFunc(r)
}

var (
	client  = &HTTPClientMock{}
	service = NewAWSService(client)
)

func TestListAccounts(t *testing.T) {
	tt := []struct {
		Body       string
		StatusCode int

		Result       accounts
		ErrorMessage string
	}{
		{
			Body: `{
					"paginationToken": null,
					"result":
						[
							{
								"id": "id1",
								"name": "name1",
								"description": "",
								"applicationId": "app-1",
								"applicationName": "AWS Account",
								"icon": "1.png",
								"searchMetadata": {
									"AccountEmail": "a@b.com",
									"AccountId": "account1",
									"AccountName": "PROD"
								}
							},
							{
								"id": "id2",
								"name": "name2",
								"description": "",
								"applicationId": "app-1",
								"applicationName": "AWS Account",
								"icon": "2.png",
								"searchMetadata": {
									"AccountEmail": "a@b.com",
									"AccountId": "account2",
									"AccountName": "Non-PROD"
								}
							}
						]
					}`,
			StatusCode: 200,
			Result: accounts{
				Result: []account{
					{
						Id:   "id1",
						Name: "name1",
					},
					{
						Id:   "id2",
						Name: "name2",
					},
				},
			},
			ErrorMessage: "",
		},
		{
			Body:         `{"message":"Internal error","__type":"com.amazonaws.switchboard.portal#NullPointerException"}`,
			StatusCode:   200,
			ErrorMessage: "operation failed due to: Internal error",
		},
	}

	for _, test := range tt {
		client.DoFunc = func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				Body:       io.NopCloser(strings.NewReader(test.Body)),
				StatusCode: test.StatusCode,
			}, nil
		}

		p, err := service.getAccounts("url", "token")
		if err != nil && err.Error() != test.ErrorMessage {
			t.Fatalf("want %s, got %s", test.ErrorMessage, err.Error())
		}

		if len(test.ErrorMessage) == 0 && !reflect.DeepEqual(p, test.Result) {
			t.Fatalf("want %v, got %v", test.Result, p)
		}
	}
}

func TestListProfiles(t *testing.T) {
	tt := []struct {
		Body       string
		StatusCode int

		Result       profiles
		ErrorMessage string
	}{
		{
			Body: `{
					"paginationToken": null,
					"result":
						[
							{
								"id": "id1",
								"name": "name1",
								"description": "",
								"url": "url1",
								"protocol": "SAML",
								"relayState": null
							},
							{
								"id": "id2",
								"name": "name2",
								"description": "",
								"url": "url2",
								"protocol": "SAML",
								"relayState": null
							}
						]
					}`,
			StatusCode: 200,
			Result: profiles{
				Result: []profile{
					{
						Name: "name1",
					},
					{
						Name: "name2",
					},
				},
			},
			ErrorMessage: "",
		},
		{
			Body:         `{"message":"Instance id 123 format is incorrect","__type":"com.amazonaws.switchboard.portal#InvalidRequestException"}`,
			StatusCode:   200,
			ErrorMessage: "operation failed due to: Instance id 123 format is incorrect",
		},
	}

	for _, test := range tt {
		client.DoFunc = func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				Body:       io.NopCloser(strings.NewReader(test.Body)),
				StatusCode: test.StatusCode,
			}, nil
		}

		p, err := service.getProfiles("url", "token")
		if err != nil && err.Error() != test.ErrorMessage {
			t.Fatalf("want %s, got %s", test.ErrorMessage, err.Error())
		}

		if len(test.ErrorMessage) == 0 && !reflect.DeepEqual(p, test.Result) {
			t.Fatalf("want %v, got %v", test.Result, p)
		}
	}
}

func TestGetCredentials(t *testing.T) {
	ts := []struct {
		Body       string
		StatusCode int

		Result       credentials
		ErrorMessage string
	}{
		{
			Body:       `{"roleCredentials":{"accessKeyId":"ASI","secretAccessKey":"s56dMibME","sessionToken":"IQoJb3JpZX2VjEP","expiration":1639987624000}}`,
			StatusCode: 200,
			Result: credentials{
				RoleCredentials: roleCredentials{
					AccessKeyId:     "ASI",
					SecretAccessKey: "s56dMibME",
					SessionToken:    "IQoJb3JpZX2VjEP",
					Expiration:      1639987624000,
				},
			},
			ErrorMessage: "",
		},
		{
			Body:         `{"message":"No access","__type":"com.amazonaws.switchboard.portal#ForbiddenException"}`,
			StatusCode:   200,
			ErrorMessage: "operation failed due to: No access",
		},
		{
			Body:         `{"message":"Internal error","__type":"com.amazonaws.switchboard.portal#NullPointerException"}`,
			StatusCode:   200,
			ErrorMessage: "operation failed due to: Internal error",
		},
	}

	for _, test := range ts {
		client.DoFunc = func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				Body:       io.NopCloser(strings.NewReader(test.Body)),
				StatusCode: test.StatusCode,
			}, nil
		}

		p, err := service.getCredentials("url", "token")
		if err != nil && err.Error() != test.ErrorMessage {
			t.Fatalf("want %s, got %s", test.ErrorMessage, err.Error())
		}

		if len(test.ErrorMessage) == 0 && !reflect.DeepEqual(p, test.Result) {
			t.Fatalf("want %v, got %v", test.Result, p)
		}
	}
}

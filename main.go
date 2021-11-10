package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/urfave/cli/v2"
)

var app = cli.NewApp()

func main() {
	info()
	commands()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func info() {
	app.Name = "aws-one-punch"
	app.Usage = "one punch to access AWS resoruces in command line"
	app.Version = "1.0.0"
}

func commands() {
	domain, err := getDomain()
	if err != nil {
		log.Fatal(err)
	}
	app.Commands = []*cli.Command{
		{
			Name:    "list-accounts",
			Aliases: []string{"ls-a"},
			Usage:   "List accounts",
			Action: func(c *cli.Context) error {
				token, err := GetAwsSsoToken(domain)
				if err != nil {
					return err
				}
				accounts, err := getAccounts("https://portal.sso.ap-southeast-2.amazonaws.com/instance/appinstances", token)
				if err != nil {
					return err
				}
				if len(accounts.Result) > 0 {
					for i := 0; i < len(accounts.Result); i++ {
						fmt.Printf("AccountId: %s, accountName: %s\n", accounts.Result[i].Id, accounts.Result[i].Name)
					}
				}
				return nil
			},
		},
		{
			Name:    "list-profiles",
			Aliases: []string{"ls-p"},
			Usage:   "List profiles under an account",
			Action: func(c *cli.Context) error {
				accountId := c.Args().Get(0)
				token, err := GetAwsSsoToken(domain)
				if err != nil {
					log.Fatalln(err)
				}
				profiles, err := getProfiles(fmt.Sprintf("https://portal.sso.ap-southeast-2.amazonaws.com/instance/appinstance/%s/profiles", accountId), token)
				if err != nil {
					log.Fatalln(err)
				}
				if len(profiles.Result) > 0 {
					for i := 0; i < len(profiles.Result); i++ {
						fmt.Printf("ProfileName: %s\n", profiles.Result[i].Name)
					}
					return nil
				}
				fmt.Printf("no profiles found for account %s\n", accountId)
				return nil
			},
		},
		{
			Name:    "access",
			Aliases: []string{"access"},
			Usage:   "Access AWS Resource with a profile",
			Action: func(c *cli.Context) error {
				accountId := c.Args().Get(0)
				profileName := c.Args().Get(1)
				token, err := GetAwsSsoToken(domain)
				if err != nil {
					log.Fatalln(err)
				}
				cs, err := getCredentials(fmt.Sprintf("https://portal.sso.ap-southeast-2.amazonaws.com/federation/credentials/?account_id=%s&role_name=%s&debug=true", accountId, profileName), token)
				if err != nil {
					log.Fatalln(err)
				}
				err = updateCredentialFile(cs)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Printf("AWS access granted with account %s and profile %s\n", accountId, profileName)
				return nil
			},
		},
	}
}

func getDomain() (string, error) {
	byts, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return "", err
	}

	config := make(map[string]string)
	err = json.Unmarshal(byts, &config)
	if err != nil {
		return "", err
	}

	doamin, exist := config["aws-console-domain"]
	if !exist {
		return "", fmt.Errorf("invaid aws-console-domain configured")
	}

	return doamin, nil
}

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
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&c)
	if err != nil {
		log.Fatalln(err)
	}

	return c, nil
}

func updateCredentialFile(c credentials) error {
	usr, _ := user.Current()
	filePath := fmt.Sprintf("%s/.aws/credentials", usr.HomeDir)
	// remove the credentials file first
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to remove existing credentials file %s", err.Error())
	}

	// create the credentials file
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create credentials file %s", err.Error())
	}
	defer f.Close()

	// update the credentials file
	content := fmt.Sprintf("[default]\naws_access_key_id=%s\naws_secret_access_key=%s\naws_session_token=%s", c.RoleCredentials.AccessKeyId, c.RoleCredentials.SecretAccessKey, c.RoleCredentials.SessionToken)
	_, err = f.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to update the credentials file %s", err.Error())
	}

	return nil
}

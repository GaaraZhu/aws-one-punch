package main

import (
	"errors"
	"fmt"
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
	app.Usage = "one punch to grant all command line windows AWS access in MacOS"
	app.Version = "1.0.0"
}

func commands() {
	domain, err := getDomain()
	if err != nil {
		log.Fatal(err)
	}
	awsService := NewAWSService(&http.Client{
		Timeout: time.Second * 20,
	})
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
				accounts, err := awsService.getAccounts("https://portal.sso.ap-southeast-2.amazonaws.com/instance/appinstances", token)
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
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "account-id", Required: true},
			},
			Action: func(c *cli.Context) error {
				accountId := c.Value("account-id")
				token, err := GetAwsSsoToken(domain)
				if err != nil {
					log.Fatalln(err)
				}
				profiles, err := awsService.getProfiles(fmt.Sprintf("https://portal.sso.ap-southeast-2.amazonaws.com/instance/appinstance/%s/profiles", accountId), token)
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
			Aliases: []string{"a"},
			Usage:   "Access AWS Resource with a profile",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "account-name", Required: true},
				&cli.StringFlag{Name: "profile-name", Required: true},
			},
			Action: func(c *cli.Context) error {
				accountId := c.Value("account-name")
				profileName := c.Value("profile-name")
				token, err := GetAwsSsoToken(domain)
				if err != nil {
					log.Fatalln(err)
				}
				cs, err := awsService.getCredentials(fmt.Sprintf("https://portal.sso.ap-southeast-2.amazonaws.com/federation/credentials/?account_id=%s&role_name=%s&debug=true", accountId, profileName), token)
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
	doamin := os.Getenv("AWS_CONSOLE_DOMAIN")
	if len(doamin) == 0 {
		return "", fmt.Errorf("invaid AWS_CONSOLE_DOMAIN configured")
	}

	return doamin, nil
}

func updateCredentialFile(c credentials) error {
	usr, _ := user.Current()
	folderPath := fmt.Sprintf("%s/.aws", usr.HomeDir)
	if awsFolderExists, _ := pathExists(folderPath); !awsFolderExists {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create aws folder %s due to %s", folderPath, err.Error())
		}
	}

	filePath := fmt.Sprintf("%s/credentials", folderPath)
	// remove credential file if it exists
	if exists, _ := pathExists(filePath); exists {
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to remove existing credentials file %s", err.Error())
		}
	}

	// create the credentials file
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create credentials file %s due to %s", filePath, err.Error())
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

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

package main

import (
	"golang.org/x/crypto/pbkdf2"

	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	salt       = "saltysalt"
	iv         = "                "
	length     = 16
	password   = ""
	iterations = 1003
)

type Cookie struct {
	Domain         string
	Key            string
	Value          string
	EncryptedValue []byte
	Expires_utc    int64
}

func (c *Cookie) DecryptedValue() string {
	if c.Value > "" {
		return c.Value
	}

	if len(c.EncryptedValue) > 0 {
		encryptedValue := c.EncryptedValue[3:]
		return decryptValue(encryptedValue)
	}

	return ""
}

// Retry until the SSO token is available due to a delay up to 30 seconds after the SSO authentication
// before it's updated in the Cookie due to Chrome's persistence implementation with SQLitePersistentCookieStore
// (https://www.chromium.org/developers/design-documents/network-stack/cookiemonster/).
func GetAwsSsoTokenWithRetry(domain string) (string, error) {
	for i := 0; i <= 60; i++ { // Retry for maximum 5 minutes - 5 seconds backoff and 60 maximum attempts
		token, err := getAwsSsoToken(domain)
		if err != nil {
			fmt.Printf("Warning: %s. If you have already done this, please wait for aws-one-punch to execute the command.\n", err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		return token, nil
	}

	return "", fmt.Errorf("failed to obtain AWS SSO token for domain %s", domain)
}

func getAwsSsoToken(domain string) (string, error) {
	// password will be used to decrypt the encrypted cookie value
	if password == "" {
		password = getPassword()
	}
	for _, cookie := range getCookies(domain) {
		if cookie.Key == "x-amz-sso_authn" {
			// if cookie.Expires_utc
			currentTime := time.Now().UTC().UnixNano() / int64(time.Second)
			if currentTime <= cookie.Expires_utc {
				return cookie.DecryptedValue(), nil
			} else {
				return "", fmt.Errorf("AWS SSO Token expired, please finish the SSO in the user portal first: https://%s/start/#/", domain)
			}
		}
	}
	return "", fmt.Errorf("no AWS SSO token found, please finish the SSO in the user portal first: https://%s/start/#/", domain)
}

func decryptValue(encryptedValue []byte) string {
	key := pbkdf2.Key([]byte(password), []byte(salt), iterations, length, sha1.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	decrypted := make([]byte, len(encryptedValue))
	cbc := cipher.NewCBCDecrypter(block, []byte(iv))
	cbc.CryptBlocks(decrypted, encryptedValue)

	plainText, err := aesStripPadding(decrypted)
	if err != nil {
		fmt.Println("Error decrypting:", err)
		return ""
	}
	return string(plainText)
}

// In the padding scheme the last <padding length> bytes
// have a value equal to the padding length, always in (1,16]
func aesStripPadding(data []byte) ([]byte, error) {
	if len(data)%length != 0 {
		return nil, fmt.Errorf("decrypted data block length is not a multiple of %d", length)
	}
	paddingLen := int(data[len(data)-1])
	if paddingLen > 16 {
		return nil, fmt.Errorf("invalid last block padding length: %d", paddingLen)
	}
	return data[:len(data)-paddingLen], nil
}

func getPassword() string {
	parts := strings.Fields("security find-generic-password -wga Chrome")
	cmd := parts[0]
	parts = parts[1:]

	out, err := exec.Command(cmd, parts...).Output()
	if err != nil {
		log.Fatal("error finding password ", err)
	}

	return strings.Trim(string(out), "\n")
}

func getCookies(domain string) (cookies []Cookie) {
	usr, _ := user.Current()
	cookiesFile := fmt.Sprintf("%s/Library/Application Support/Google/Chrome/Default/Cookies", usr.HomeDir)

	db, err := sql.Open("sqlite3", cookiesFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT name, value, host_key, encrypted_value, expires_utc FROM cookies WHERE host_key like ?", fmt.Sprintf("%%%s%%", domain))
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	for rows.Next() {
		var name, value, hostKey string
		var expires int64
		var encryptedValue []byte
		rows.Scan(&name, &value, &hostKey, &encryptedValue, &expires)
		cookies = append(cookies, Cookie{hostKey, name, value, encryptedValue, (expires/1000000 - 11644473600)})
	}

	return
}

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/AbeEstrada/tuit/config"
	"github.com/AbeEstrada/tuit/constants"
	"github.com/mattn/go-mastodon"
)

func SetupAuth() {
	var server string
	fmt.Print("Enter the URL of your Mastodon server: ")
	fmt.Scanln(&server)
	server = strings.TrimSpace(server)
	if !strings.HasPrefix(server, "https://") {
		server = "https://" + server
	}

	appConfig := &mastodon.AppConfig{
		ClientName:   constants.AppName,
		Website:      constants.AppUrl,
		Server:       server,
		Scopes:       "read write",
		RedirectURIs: "urn:ietf:wg:oauth:2.0:oob",
	}

	app, err := mastodon.RegisterApp(context.Background(), appConfig)
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse(app.AuthURI)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Open your browser and copy/paste the given authorization code:")
	fmt.Println(u)

	var authCode string
	fmt.Print("Paste the code here: ")
	fmt.Scanln(&authCode)

	mastodonConfig := &mastodon.Config{
		Server:       server,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
	}

	client := mastodon.NewClient(mastodonConfig)
	err = client.GetUserAccessToken(context.Background(), authCode, app.RedirectURI)
	if err != nil {
		log.Fatal(err)
	}

	newConfig := config.Config{
		Auth: config.ConfigAuth{
			Server:       client.Config.Server,
			ClientID:     client.Config.ClientID,
			ClientSecret: client.Config.ClientSecret,
			AccessToken:  client.Config.AccessToken,
		},
	}
	jsonData, err := json.MarshalIndent(newConfig, "", "    ")
	if err != nil {
		fmt.Printf("Error marshalling data: %v\n", err)
		return
	}
	configDir := config.GetConfigDir()
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory %s: %v\n", configDir, err)
		return
	}
	configFile := config.GetConfigFile()
	err = os.WriteFile(configFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
}

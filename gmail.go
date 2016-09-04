// Package gmail allows you to send email messages using Google GMail.
//
// You need to register the app on the Google server and get the configuration
// file that will be used for authorization. When you first initialize the
// application in the console will display the URL you need to go and get the
// authorization code. This code must be entered in response to the application
// and execution will continue. This function must be performed once the
// authorization keys stored in files.
//
// Registration procedure
//
// 1. Use wizard <https://console.developers.google.com/start/api?id=gmail> to
// create or select a project in the Google Developers Console and automatically
// turn on the API. Click Continue, then Go to credentials.
//
// 2. On the Add credentials to your project page, click the Cancel button.
//
// 3. At the top of the page, select the OAuth consent screen tab. Select an
// Email address, enter a Product name if not already set, and click the Save
// button.
//
// 4. Select the Credentials tab, click the Create credentials button and select
// OAuth client ID.
//
// 5. Select the application type Other, enter the name "Gmail API", and click
// the Create button.
//
// 6. Click OK to dismiss the resulting dialog.
//
// 7. Click the Download JSON button to the right of the client ID.
//
// 8. Move this file to your working directory and rename it client_secret.json.
package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// The message text that will be displayed to request the authorization code.
var RequestMessage = "Go to the following link in your browser then type the authorization code:"

// Pointer to the initialized service.
var gmailService *gmail.Service

// Init initialise GMail service. The initialization process reads a
// configuration file that must be created and received through the console of
// Google Application. Read more about the procedure of obtaining the key file
// can be read on Google: (https://developers.google.com/gmail/api/quickstart/go).
//
// When you first start the console will display a string with the URL you need
// to go and get the key to authorize the application. This key must then be
// entered immediately into the app.
//
// The resulting authorization token will be saved to the parameter file for
// future use.
func Init(config, token string) error {
	b, err := ioutil.ReadFile(config)
	if err != nil {
		return err
	}
	cfg, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return err
	}
	var oauthToken = new(oauth2.Token)
	file, err := os.Open(token)
	if err == nil {
		err = json.NewDecoder(file).Decode(oauthToken)
		file.Close()
	}
	if err != nil {
		authURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		fmt.Print(RequestMessage, '\n', authURL, '\n')
		var code string
		if _, err = fmt.Scan(&code); err != nil {
			return err
		}
		oauthToken, err = cfg.Exchange(oauth2.NoContext, code)
		if err != nil {
			return err
		}
		file, err = os.Create(token)
		if err != nil {
			return err
		}
		if err = json.NewEncoder(file).Encode(oauthToken); err != nil {
			file.Close()
			return err
		}
		file.Close()
	}
	gmailService, err = gmail.New(cfg.Client(context.Background(), oauthToken))
	return err
}

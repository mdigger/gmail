# Gmail sender

[![GoDoc](https://godoc.org/github.com/mdigger/gmail?status.svg)](https://godoc.org/github.com/mdigger/gmail)
[![Build Status](https://travis-ci.org/mdigger/gmail.svg?branch=master)](https://travis-ci.org/mdigger/gmail)
[![Coverage Status](https://coveralls.io/repos/github/mdigger/gmail/badge.svg)](https://coveralls.io/github/mdigger/gmail?branch=master)

Library to send messages using Google GMail.

You need to register the app on the Google server and get the configuration file that will be used for authorization. When you first initialize the application in the console will display the URL you need to go and get the authorization code. This code must be entered in response to the application and execution will continue. This function must be performed once the authorization keys stored in files.

### Message send example

```go
package main

import (
	"io/ioutil"
	"log"

	"github.com/mdigger/gmail"
)

func main() {
	// initialize the library
	if err := gmail.Init("config.json", "token.json"); err != nil {
		log.Fatal(err)
	}
	// create a new message
	msg, err := gmail.NewMessage(
		"Subject", // subject
		"me",      // from
		[]string{"Test User <test@example.com>"}, // to
		nil, // cc
		[]byte(`<html><p>body text</p></html>`), // message body
	)
	if err != nil {
		log.Fatal(err)
	}
	// attach file
	if err = msg.AddFile("README.md"); err != nil {
		log.Fatal(err)
	}
	// send a message
	if err := msg.Send(); err != nil {
		log.Fatal(err)
	}
}
```

### Registration procedure

- Use [this wizard](https://console.developers.google.com/start/api?id=gmail) to create or select a project in the Google Developers Console and automatically turn on the API. Click **Continue**, then **Go to credentials**.
- On the **Add credentials to your project** page, click the **Cancel** button.
- At the top of the page, select the **OAuth consent screen** tab. Select an **Email address**, enter a **Product name** if not already set, and click the **Save** button.
- Select the **Credentials** tab, click the **Create credentials** button and select **OAuth client ID**.
- Select the application type **Other**, enter the name "Gmail API", and click the **Create** button.
- Click **OK** to dismiss the resulting dialog.
- Click the **Download JSON** button to the right of the client ID.
- Move this file to your working directory and rename it `client_secret.json`.
package gmail_test

import (
	"log"

	"github.com/mdigger/gmail"
)

func Example() {
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

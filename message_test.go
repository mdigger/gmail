package gmail

import (
	"bytes"
	"testing"
)

func TestMessageText(t *testing.T) {
	msg, err := NewMessage(
		"Subject",
		"Dmitrys <dmitrys@xyzrd.com>",
		[]string{
			"I am<sedykh@gmail.com>",
			"I am too <dmitrys@xyzrd.com>",
			"d3@yandex.ru"},
		nil)
	if err != nil {
		t.Fatal(err)
	}
	// err = msg.Body([]byte(`text body`))
	// err = msg.Body([]byte(`<p>html body</p>`))
	err = msg.Body(html)
	if err != nil {
		t.Fatal(err)
	}
	// err = msg.File("test_file.html", html)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// err = msg.File("test_file.txt", text)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// pretty.Println(msg)

	var buf bytes.Buffer
	if err := msg.writeTo(&buf); err != nil {
		t.Fatal(err)
	}
	print(buf.String())
}

func TestMessageSend(t *testing.T) {
	if err := Init("config.json", "token.json"); err != nil {
		t.Fatal(err)
	}
	msg, err := NewMessage(
		"Тестовое сообщение",
		"",
		[]string{"Дмитрий Седых <d3@yandex.ru>"},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	// msg.Body(text)
	err = msg.Body(html)
	if err != nil {
		t.Fatal(err)
	}
	err = msg.File("test_file.html", html)
	if err != nil {
		t.Fatal(err)
	}
	err = msg.File("test_file.txt", text)
	if err != nil {
		t.Fatal(err)
	}

	// pretty.Println(msg)

	if err := msg.Send(); err != nil {
		t.Fatal(err)
	}
}

var (
	text = []byte(`Message body

This is a text text message.
`)
	html = []byte(`<html><h2>Message body</h2>

<p>This is a html text message.</p></html>`)
)

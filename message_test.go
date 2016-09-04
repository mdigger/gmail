package gmail

import (
	"bytes"
	"io/ioutil"
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
		[]string{"Дмитрий Седых <d3@yandex.ru>"},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	// err = msg.Body([]byte(`text body`))
	// err = msg.Body([]byte(`<p>html body</p>`))
	err = msg.SetBody(html)
	if err != nil {
		t.Error(err)
	}
	err = msg.Attach("test_file.html", html)
	if err != nil {
		t.Error(err)
	}
	err = msg.Attach("test_file.txt", text)
	if err != nil {
		t.Error(err)
	}
	err = msg.Attach("test_file.txt", nil)
	if err != nil {
		t.Error(err)
	}
	if msg.Has("test_file.txt") {
		t.Error("bad delete attachment")
	}
	err = msg.Attach("test_file.bin", bin)
	if err != nil {
		t.Error(err)
	}

	// pretty.Println(msg)

	var buf bytes.Buffer
	if err := msg.writeTo(&buf); err != nil {
		t.Fatal(err)
	}
	print(buf.String())
}

func TestBadMessage(t *testing.T) {
	if _, err := NewMessage("", "me", []string{"bad"}, nil, nil); err == nil {
		t.Error(err)
	}
	if _, err := NewMessage("", "me", nil, []string{"bad"}, nil); err == nil {
		t.Error(err)
	}
	if _, err := NewMessage("", "bad", nil, nil, nil); err == nil {
		t.Error("bad from email")
	}
	if _, err := NewMessage("", "me", []string{"d3@ya.ru"}, nil, bin); err == nil {
		t.Error("bad message text")
	}
	if _, err := NewMessage("", "me", nil, nil, nil); err == nil {
		t.Error("bad recipient")
	}
}

func TestBadAttach(t *testing.T) {
	msg, err := NewMessage("", "", []string{"d3@yandex.ru"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := msg.Attach("", html); err == nil {
		t.Error("bad file name")
	}
	if err := msg.writeTo(ioutil.Discard); err == nil {
		t.Error("content undefined")
	}
	if err := msg.Attach("~/test/file.name", html); err != nil {
		t.Error(err)
	}
	if err := msg.writeTo(ioutil.Discard); err != nil {
		t.Error(err)
	}
	if err := msg.Attach("..", text); err == nil {
		t.Error("bad file name 2")
	}
	if err := msg.Attach("~/test/file.name", text); err != nil {
		t.Error(err)
	}
	if err := msg.Attach("~/test/file.name", bin); err != nil {
		t.Error(err)
	}

	// pretty.Println(msg)
}

func TestSimpleMessage(t *testing.T) {
	msg, err := NewMessage(
		"Subject",
		"",
		[]string{"Дмитрий Седых <d3@yandex.ru>"},
		nil,
		html,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := msg.writeTo(ioutil.Discard); err != nil {
		t.Error(err)
	}
	gmailService = nil
	if msg.Send() == nil {
		t.Error("not initialized service")
	}
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
		html,
	)
	if err != nil {
		t.Fatal(err)
	}
	err = msg.Attach("test_file.html", html)
	if err != nil {
		t.Error(err)
	}
	err = msg.Attach("test_file.txt", text)
	if err != nil {
		t.Error(err)
	}

	// pretty.Println(msg)

	if err := msg.Send(); err != nil {
		t.Fatal(err)
	}
}

var (
	text = []byte("Message body\nThis is a text text message.")
	html = []byte(`<html>
<h2>Message body</h2>
<p>This is a html text message.</p>
</html>`)
	bin = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
)

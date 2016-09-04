package gmail_test

import (
	"io/ioutil"
	"log"

	"github.com/mdigger/gmail"
)

func Example() {
	// инициализируем библиотеку
	if err := gmail.Init("config.json", "token.json"); err != nil {
		log.Fatal(err)
	}
	// создаем новое сообщение
	msg, err := gmail.NewMessage(
		"Subject",                                // тема сообщения
		"sender@example.com",                     // от кого
		[]string{"Test User <test@example.com>"}, // кому
		nil, // копия
	)
	if err != nil {
		log.Fatal(err)
	}
	// задаем текст письма в формате HTML или текст
	msg.Body([]byte(`<html><p>body text</p></html>`))
	// присоединяем файл
	var filename = "README.md"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	msg.File(filename, data)
	// отправляем сообщение
	if err := msg.Send(); err != nil {
		log.Fatal(err)
	}
}

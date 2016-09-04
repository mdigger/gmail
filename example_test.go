package gmail_test

import (
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
		"Subject", // тема сообщения
		"me",      // от кого
		[]string{"Test User <test@example.com>"}, // кому
		nil, // копия
		[]byte(`<html><p>body text</p></html>`), // текст сообщения
	)
	if err != nil {
		log.Fatal(err)
	}
	// присоединяем файл
	if err = msg.AddFile("README.md"); err != nil {
		log.Fatal(err)
	}
	// отправляем сообщение
	if err := msg.Send(); err != nil {
		log.Fatal(err)
	}
}

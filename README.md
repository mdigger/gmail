# Gmail sender

[![GoDoc](https://godoc.org/github.com/mdigger/gmail?status.svg)](https://godoc.org/github.com/mdigger/gmail)
[![Build Status](https://travis-ci.org/mdigger/gmail.svg?branch=master)](https://travis-ci.org/mdigger/gmail)
[![Coverage Status](https://coveralls.io/repos/github/mdigger/gmail/badge.svg)](https://coveralls.io/github/mdigger/gmail?branch=master)

Библиотека для отправки сообщений через Google GMail.

Для работы библиотеки необходимо зарегистрировать приложение на сервере Google и получить конфигурационный файл, который будет использоваться для авторизации. При первой инициализации приложения в консоль будет выведен URL, по которому необходимо перейти и получить авторизационный код. Этот код нужно будет ввести в ответ приложению и выполнение продолжится. Данную функцию необходимо выполнить один раз: в дальнейшем ключи авторизации сохранятся в файлах.

### Пример отправки сообщения

	package main

	import (
		"io/ioutil"
		"log"

		"github.com/mdigger/gmail"
	)

	func main() {
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
		var filename = "README.md"
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		if err = msg.File(filename, data); err != nil {
			log.Fatal(err)
		}
		// отправляем сообщение
		if err := msg.Send(); err != nil {
			log.Fatal(err)
		}
	}


### Подробнее о процедуре регистрации

- Use [this wizard](https://console.developers.google.com/start/api?id=gmail) to create or select a project in the Google Developers Console and automatically turn on the API. Click **Continue**, then **Go to credentials**.
- On the **Add credentials to your project** page, click the **Cancel** button.
- At the top of the page, select the **OAuth consent screen** tab. Select an **Email address**, enter a **Product name** if not already set, and click the **Save** button.
- Select the **Credentials** tab, click the **Create credentials** button and select **OAuth client ID**.
- Select the application type **Other**, enter the name "Gmail API", and click the **Create** button.
- Click **OK** to dismiss the resulting dialog.
- Click the **Download JSON** button to the right of the client ID.
- Move this file to your working directory and rename it `client_secret.json`.
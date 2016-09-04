// Package gmail позволяет отправлять почтовые сообщения, используя Google
// GMail.
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

// Текст сообщения, которое будет выведено для запроса кода авторизации.
var RequestMessage = "Go to the following link in your browser then type the authorization code:"

// Указатель на инициализированный сервис GMail.
var gmailService *gmail.Service

// Init инициализирует сервис GMail. В процессе инициализации читается файл
// конфигурации, который должен быть создан и получен через консоль Google
// Application. Подробнее о процедуре создания и получения файла с ключами
// можно прочитать на сервер Google:
// (https://developers.google.com/gmail/api/quickstart/go).
//
// При первом запуске в консоль будет выведена строка с URL, по которому нужно
// перейти и получить ключ для авторизации приложения. Этот ключ затем должен
// быть введен тут же в приложение.
//
// Полученный в результате авторизации токен будет сохранен в указанном
// в параметре файле для будущего использования.
func Init(config, token string) error {
	// читаем конфигурационный файл для авторизации
	b, err := ioutil.ReadFile(config)
	if err != nil {
		return err
	}
	// инициализируем сервис (gmail.MailGoogleComScope)
	cfg, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return err
	}
	// инициализируем токен
	var oauthToken = new(oauth2.Token)
	// читаем содержимое файла для авторизации
	file, err := os.Open(token)
	if err == nil { // разбираем содержимое файла в токен
		err = json.NewDecoder(file).Decode(oauthToken)
		file.Close()
	}
	// если файла нет или не удалось его разобрать, то необходимо его
	// запросить с сервера.
	if err != nil {
		// формируем строку для запроса кода и выводим ее в консоль
		authURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		fmt.Print(RequestMessage, '\n', authURL, '\n')
		// читаем из консоли полученный пользователем код
		var code string
		if _, err = fmt.Scan(&code); err != nil {
			return err
		}
		// получаем авторизационный токен
		oauthToken, err = cfg.Exchange(oauth2.NoContext, code)
		if err != nil {
			return err
		}
		// создаем файл с авторизационной информацией
		file, err = os.Create(token)
		if err != nil {
			return err
		}
		// сохраняем в нем содержимое токена
		if err = json.NewEncoder(file).Encode(oauthToken); err != nil {
			file.Close()
			return err
		}
		file.Close()
	}
	// инициализируем сервис для доступа к GMail
	gmailService, err = gmail.New(cfg.Client(context.Background(), oauthToken))
	return err
}

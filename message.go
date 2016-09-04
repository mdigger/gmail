package gmail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/http"
	"net/mail"
	"net/textproto"
	"path"
	"sort"
	"strings"

	"github.com/kr/pretty"
	"google.golang.org/api/gmail/v1"
)

// Предопределенные ошибки, возвращаемые при попытке отослать сообщение.
var (
	// возвращается при попытке отослать пустое сообщение, без присоединенных
	// файлов и без текста самого сообщения
	ErrNoBody = errors.New("contents are undefined")
	// ошибка не инициализированного сервиса GMail
	ErrServiceNotInitialized = errors.New("gmail service not initialized")
)

// Message описывает почтовое сообщение.
type Message struct {
	header textproto.MIMEHeader // заголовки
	parts  map[string]*part     // список файлов по именам
}

// NewMessage формирует новое почтовое сообщение для отправки.
func NewMessage(subject, from string, to, cc []string) (*Message, error) {
	var h = make(textproto.MIMEHeader)
	// добавляем адрес от кого сообщение
	if mfrom, err := mail.ParseAddress(from); err == nil {
		from := mfrom.String()
		h.Set("From", from)
		h.Set("Reply-To", from)
	} else if err.Error() != "mail: no address" {
		return nil, fmt.Errorf("from %v", err)
	}
	// добавляем адреса кому
	if len(to) > 0 {
		if addr, err := addrsList(to); err == nil {
			h.Set("To", addr)
		} else if err.Error() != "mail: no address" {
			return nil, fmt.Errorf("to %v", err)
		}
	}
	// добавляем адреса копии
	if len(cc) > 0 {
		if addr, err := addrsList(cc); err == nil {
			h.Set("Сс", addr)
		} else if err.Error() != "mail: no address" {
			return nil, fmt.Errorf("cc %v", err)
		}
	}
	// проверяем, что хотя бы одни адрес установлен
	if h.Get("To") == "" && h.Get("Cc") == "" {
		pretty.Println(h)
		return nil, errors.New("no recipient specified")
	}
	// добавляем тему сообщения
	if subject != "" {
		h.Set("Subject", mime.QEncoding.Encode("utf-8", subject))
	}
	return &Message{header: h}, nil
}

const _body = "\000body" // имя файла с содержимым сообщения

// File присоединяет к сообщению новый файл. Передача пустого содержимого
// файла удалить файл с таким же именем, если он раньше был добавлен.
func (m *Message) File(name string, data []byte) error {
	if len(data) == 0 {
		delete(m.parts, name)
		return nil
	}
	// нормализуем имя файла, удаляя возможные пути
	if name = path.Base(name); name == "." {
		return errors.New("bad file name")
	}
	// формируем заголовок
	var h = make(textproto.MIMEHeader)
	// определяем тип содержимого файла
	var contentType = mime.TypeByExtension(path.Ext(name))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	// выбираем тип кодирования на основе типа содержимого
	var coding = "quoted-printable"
	if !strings.HasPrefix(contentType, "text") {
		// проверяем, что содержимое сообщения текстовое
		if name == _body {
			return fmt.Errorf("unsupported body content type: %v", contentType)
		}
		coding = "base64"
	}
	h.Set("Content-Transfer-Encoding", coding)
	// тип присоединения файла
	if name != _body {
		disposition := fmt.Sprintf("attachment; filename=%s", name)
		h.Set("Content-Disposition", disposition)
	}
	// сохраняем файл под его именем в контексте сообщения
	if m.parts == nil {
		m.parts = make(map[string]*part)
	}
	m.parts[name] = &part{
		header: h,
		data:   data,
	}
	return nil
}

// Body добавляет в почтовое сообщение текст. Повторный вызов данной функции
// приведет у перезаписи теста сообщения. Для сброса текста сообщения можно
// передать пустые данные.
//
// Текст должен быть в формате text/plain или text/html (определяется
// автоматически). Чтобы установить формат HTML, обрамите текст сообщения
// тегом <html>.
//
// Если сообщение в формате HTML, то его текстовое представление добавляется
// автоматически.
func (m *Message) Body(data []byte) error {
	return m.File(_body, data)
}

// Has возвращает true, если файл с таким именем зарегистрирован в сообщении
// в виде вложения.
func (m *Message) Has(name string) bool {
	_, ok := m.parts[name]
	return ok
}

// writeTo формирует и записывает текстовое представление постового сообщения.
func (m *Message) writeTo(w io.Writer) error {
	if len(m.parts) == 0 {
		return ErrNoBody // ничего нет
	}
	var h = make(textproto.MIMEHeader)
	h.Set("MIME-Version", "1.0")
	h.Set("X-Mailer", "REST GMailer (github.com/mdigger/gmail)")
	// копируем основной заголовок сообщения
	for k, v := range m.header {
		h[k] = v
	}
	// проверяем, что определено только основное сообщение, без файлов
	if len(m.parts) == 1 && m.Has(_body) {
		body := m.parts[_body] // выбираем содержимое сообщения
		// объединяем заголовок сообщения и файла с текстом
		for k, v := range body.header {
			h[k] = v
		}
		// записываем объединенный заголовок
		if err := writeHeader(w, h); err != nil {
			return err
		}
		// записываем содержимое сообщения
		if err := body.writeData(w); err != nil {
			return err
		}
		return nil
	}
	// есть присоединенные файлы
	var mw = multipart.NewWriter(w)
	defer mw.Close()
	h.Set("Content-Type",
		fmt.Sprintf("multipart/mixed; boundary=%s", mw.Boundary()))
	// записываем объединенный заголовок
	if err := writeHeader(w, h); err != nil {
		return err
	}
	// записываем присоединенные файлы и основной текст сообщения
	for _, p := range m.parts {
		pw, err := mw.CreatePart(p.header)
		if err != nil {
			return err
		}
		if err = p.writeData(pw); err != nil {
			return err
		}
	}
	return nil
}

// Send отправляет сообщение через GMail.
//
// Перед отправкой необходимо инициализировать сервис, вызвав функцию
// gmail.Init(), которая должна выполняться до старта сервера, потому что может
// потребовать ввода кода ответа при первой инициализации сервиса.
func (m *Message) Send() error {
	// проверяем, что сервис инициализирован
	if gmailService == nil || gmailService.Users == nil {
		return ErrServiceNotInitialized
	}
	// формируем сообщение в формате mail
	var buf bytes.Buffer
	if err := m.writeTo(&buf); err != nil {
		return err
	}
	// кодируем содержимое сообщения в формат Base64
	body := base64.RawURLEncoding.EncodeToString(buf.Bytes())
	// формируем сообщение в формате GMail
	var gmailMessage = &gmail.Message{Raw: body}
	// отправляем сообщение на сервер GMail
	_, err := gmailService.Users.Messages.Send("me", gmailMessage).Do()
	return err // возвращаем статус отправки сообщения
}

// part описывает часть почтового сообщения: файл или текст сообщения.
type part struct {
	header textproto.MIMEHeader // заголовки
	data   []byte               // содержимое
}

// writeHeader записывает заголовок части сообщения.
func (p *part) writeHeader(w io.Writer) error {
	return writeHeader(w, p.header)
}

// writeData записывает содержимое файла сообщения, поддерживая заданную
// систему кодирования. На данный момент реализованы только quoted-printable и
// base64 кодировки. Для всех остальный возвращается ошибка.
func (p *part) writeData(w io.Writer) (err error) {
	switch name := p.header.Get("Content-Transfer-Encoding"); name {
	case "quoted-printable":
		enc := quotedprintable.NewWriter(w)
		_, err = enc.Write(p.data)
		enc.Close()
	case "base64":
		enc := base64.NewEncoder(base64.StdEncoding, w)
		_, err = enc.Write(p.data)
		enc.Close()
	default:
		err = fmt.Errorf("unsupported transform encoding: %v", name)
	}
	return err
}

// writeHeader записывает заголовок сообщения или файла. Ключи заголовка
// сортируются в алфавитном порядке.
func writeHeader(w io.Writer, h textproto.MIMEHeader) (err error) {
	// сортируем ключи, чтобы выводить их в одинаковом виде
	var keys = make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range h[k] {
			if _, err = fmt.Fprintf(w, "%s: %s\r\n", k, v); err != nil {
				return err
			}
		}
	}
	_, err = fmt.Fprintf(w, "\r\n") // добавляем отступ от заголовка
	return err
}

// addrsList возвращает строку с адресами, сформированными из списка адресов.
func addrsList(addrs []string) (string, error) {
	mails, err := mail.ParseAddressList(strings.Join(addrs, ", "))
	if err != nil {
		return "", err
	}
	var list = make([]string, len(mails))
	for i, addr := range mails {
		list[i] = addr.String()
	}
	return strings.Join(list, ", "), nil
}

package gmail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/http"
	"net/mail"
	"net/textproto"
	"path/filepath"
	"sort"
	"strings"

	"google.golang.org/api/gmail/v1"
)

// Predefined error returned when trying to send a message.
var (
	// is returned if you try to send a blank message with no attached files and
	// no text messages
	ErrNoBody = errors.New("contents are undefined")
	// error not initialized the GMail
	ErrServiceNotInitialized = errors.New("gmail service not initialized")
)

// The Message describes an email message.
type Message struct {
	header textproto.MIMEHeader // headers
	parts  map[string]*part     // the list of file by names
}

// NewMessage creates a new email message to send.
//
// The message is always sent on behalf of an authorized user, so that the from
// field can be empty or contain the string "me". If you set this field to a
// different address, it will appear in the Reply-To and when you reply to this
// message will use the sender's address.
//
// You can specify email address in the following formats (supported by parsing
// the name and email address):
//
//	test@example.com
//	<test@example.com>
//	TestUser <test@example.com>
//
// Parsing checks the validity of the format of the email address. You must
// specify at least one address to send (to or cc), or when trying to send
// messages will return an error.
//
// You can use text or HTML message format is determined automatically. To
// guarantee that the format will be specified as HTML, consider wrapping the
// text with <html> tag. When adding the HTML content of the message, text
// version, to support legacy mail program will be added automatically. When
// you try to add as text message binary data will return an error. You can
// set nil as a parameter to empty message body.
func NewMessage(subject, from string, to, cc []string, body []byte) (*Message, error) {
	var h = make(textproto.MIMEHeader)
	if from != "" && from != "me" {
		if mfrom, err := mail.ParseAddress(from); err == nil {
			from := mfrom.String()
			h.Set("From", from)
			h.Set("Reply-To", from)
		} else if err.Error() != "mail: no address" {
			return nil, fmt.Errorf("from %v", err)
		}
	}
	if len(to) > 0 {
		if addr, err := addrsList(to); err == nil {
			h.Set("To", addr)
		} else if err.Error() != "mail: no address" {
			return nil, fmt.Errorf("to %v", err)
		}
	}
	if len(cc) > 0 {
		if addr, err := addrsList(cc); err == nil {
			h.Set("Сс", addr)
		} else if err.Error() != "mail: no address" {
			return nil, fmt.Errorf("cc %v", err)
		}
	}
	if h.Get("To") == "" && h.Get("Cc") == "" {
		return nil, errors.New("no recipient specified")
	}
	if subject != "" {
		h.Set("Subject", mime.QEncoding.Encode("utf-8", subject))
	}
	var msg = &Message{header: h}
	if len(body) > 0 {
		if err := msg.SetBody(body); err != nil {
			return msg, err
		}
	}
	return msg, nil
}

const _body = "\000body" // the file name with the contents of the message

// Attach attaches to the message an attachment as a file. Passing an empty
// content deletes the file with the same name if it was previously added.
func (m *Message) Attach(name string, data []byte) error {
	if len(data) == 0 {
		if m.parts != nil {
			delete(m.parts, name)
		}
		return nil
	}
	name = filepath.Base(name)
	switch name {
	case ".", "..", string(filepath.Separator):
		return fmt.Errorf("bad file name: %v", name)
	}
	var h = make(textproto.MIMEHeader)
	var contentType = mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	var coding = "quoted-printable"
	if !strings.HasPrefix(contentType, "text") {
		if name == _body {
			return fmt.Errorf("unsupported body content type: %v", contentType)
		}
		coding = "base64"
	}
	h.Set("Content-Transfer-Encoding", coding)
	if name != _body {
		disposition := fmt.Sprintf("attachment; filename=%s", name)
		h.Set("Content-Disposition", disposition)
	}
	if m.parts == nil {
		m.parts = make(map[string]*part)
	}
	m.parts[name] = &part{
		header: h,
		data:   data,
	}
	return nil
}

// AddFile reads the contents of specified in the parameter file and attaches
// it as an attachment to the message.
func (m *Message) AddFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return m.Attach(filename, data)
}

// SetBody sets the contents of the text of the letter.
//
// You can use text or HTML message format (is determined automatically). To
// guarantee that the format will be specified as HTML, consider wrapping the
// text with <html> tag. When adding the HTML content, text version, to support
// legacy mail program will be added automatically. When you try to add as
// message binary data will return an error. You can pass as a parameter the nil,
// then the message will be without a text submission.
func (m *Message) SetBody(data []byte) error {
	return m.Attach(_body, data)
}

// Has returns true if a file with that name was in the message as an attachment.
func (m *Message) Has(name string) bool {
	_, ok := m.parts[name]
	return ok
}

// writeTo generates and writes the text representation of mail messages.
func (m *Message) writeTo(w io.Writer) error {
	if len(m.parts) == 0 {
		return ErrNoBody
	}
	var h = make(textproto.MIMEHeader)
	h.Set("MIME-Version", "1.0")
	h.Set("X-Mailer", "REST GMailer (github.com/mdigger/gmail)")
	// copy the primary header of the message
	for k, v := range m.header {
		h[k] = v
	}
	// check that only defined the basic message, no file
	if len(m.parts) == 1 && m.Has(_body) {
		body := m.parts[_body]
		for k, v := range body.header {
			h[k] = v
		}
		writeHeader(w, h)
		if err := body.writeData(w); err != nil {
			return err
		}
		return nil
	}
	// there are attached files
	var mw = multipart.NewWriter(w)
	defer mw.Close()
	h.Set("Content-Type",
		fmt.Sprintf("multipart/mixed; boundary=%s", mw.Boundary()))
	writeHeader(w, h)
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

// Send sends the message through GMail.
//
// Before sending, you must initialize the service by calling the Init function.
func (m *Message) Send() error {
	if gmailService == nil || gmailService.Users == nil {
		return ErrServiceNotInitialized
	}
	var buf bytes.Buffer
	m.writeTo(&buf)
	body := base64.RawURLEncoding.EncodeToString(buf.Bytes())
	var gmailMessage = &gmail.Message{Raw: body}
	_, err := gmailService.Users.Messages.Send("me", gmailMessage).Do()
	return err
}

// part describes part email message: the file or message.
type part struct {
	header textproto.MIMEHeader // headers
	data   []byte               // content
}

// writeData writes the contents of the message file with maintain the coding
// system. At the moment only implemented quoted-printable and base64 encoding.
// For all others, an error is returned.
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

// writeHeader writes the header of the message or file. The keys of the header
// are sorted alphabetically.
func writeHeader(w io.Writer, h textproto.MIMEHeader) {
	var keys = make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range h[k] {
			fmt.Fprintf(w, "%s: %s\r\n", k, v)
		}
	}
	fmt.Fprintf(w, "\r\n") // add the offset from the header
}

// addrsList returns a string with the addresses generated from the address list.
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

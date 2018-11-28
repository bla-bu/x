package x

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/txthinking/mailx"
)

// MailgunAPIURL is the base url of maingun api
const MailgunAPIURL = "https://api.mailgun.net/v3"

// Mailgun is your maingun domain, api key config
type Mailgun struct {
	Domain string
	APIKey string
}

// Send can send email message by Maingun
func (m *Mailgun) Send(msg *mailx.Message) error {
	msgr, err := msg.Reader()
	if err != nil {
		return err
	}
	to := make([]string, len(msg.To))
	for i := range msg.To {
		to[i] = msg.To[i].String()
	}
	bdry := mailx.MakeBoundary()
	body, err := MultipartFormDataFromReader(
		map[string][]string{
			"to": {strings.Join(to, ",")},
		},
		map[string][]io.Reader{
			"message": {msgr},
		},
		bdry,
	)

	tspt := &http.Transport{
		TLSClientConfig:    nil,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tspt}
	r, err := http.NewRequest("POST", MailgunAPIURL+"/"+m.Domain+"/messages.mime", body)
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "multipart/form-data; boundary="+bdry)
	r.SetBasicAuth("api", m.APIKey)

	res, err := client.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return nil
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = errors.New(bytes.NewBuffer(b).String())
	return err
}

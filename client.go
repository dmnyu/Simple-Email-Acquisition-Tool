package go_email

import (
	"bufio"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"regexp"
)

var plaintext = regexp.MustCompile("text/plain")

type EmailAccount struct {
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
}

type Email struct {
	Message *imap.Message
	Section *imap.BodySectionName
}

func GetCreds(account string) (EmailAccount, error) {
	credsMap := map[string]EmailAccount{}
	source, err := ioutil.ReadFile("/etc/go-email.yml")
	if err != nil {
		return EmailAccount{}, err
	}

	err = yaml.Unmarshal(source, &credsMap)
	if err != nil {
		return EmailAccount{}, err
	}

	for k, v := range credsMap {
		if account == k {
			return v, nil
		}
	}

	return EmailAccount{}, fmt.Errorf("Credentials file did not contain %s\n", account)
}

func GetClient(accountName string) (*client.Client, error) {
	log.Println("Connecting to server...")

	account, err := GetCreds(accountName)
	if err != nil {
		return nil, err
	}

	// Connect to server
	addr := fmt.Sprintf(account.Server + ":" + account.Port)
	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return c, err
	}
	log.Println("Connected")
	return c, nil
}

func GetMailboxes(c *client.Client) ([]string, error) {
	mailboxes := []string{}
	mailboxChannel := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxChannel)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxChannel {
		mailboxes = append(mailboxes, m.Name)
	}

	if err := <-done; err != nil {
		return mailboxes, err
	}

	return mailboxes, nil
}

func GetMailbox(c *client.Client, name string) (*imap.MailboxStatus, error) {
	mailbox, err := c.Select(name, true)
	if err != nil {
		return mailbox, err
	}
	return mailbox, nil
}

func GetMessage(i uint32, c *client.Client) (Email, error) {
	seqset := new(imap.SeqSet)
	seqset.AddNum(i)
	done := make(chan error, 1)

	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)

	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	if err := <-done; err != nil {
		return Email{}, nil
	}

	msg := <-messages

	return Email{msg, &section}, nil
}

func GetMessages(from uint32, to uint32, c *client.Client) (chan *imap.Message, error) {

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)
	messages := make(chan *imap.Message, to)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	if err := <-done; err != nil {
		return nil, err
	}

	return messages, nil
}

func WriteMessage(writer *bufio.Writer, email Email) error {
	r := email.Message.GetBody(email.Section)
	mr, err := mail.CreateReader(r)
	if err != nil {
		return err
	}

	writer.WriteString("\n")
	writer.WriteString(fmt.Sprintf("From %s %s\n", mr.Header.Get("from"), mr.Header.Get("date")))
	writer.Flush()

	fields := mr.Header.Fields()
	for fields.Next() {
		writer.WriteString(fmt.Sprintf("%s: %v\n", fields.Key(), fields.Value()))
		writer.Flush()
	}
	writer.WriteString("\n")
	for {

		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			hf := h.Fields()
			for hf.Next() {
				writer.WriteString(fmt.Sprintf("%s: %s\n", hf.Key(), hf.Value()))
				writer.Flush()
			}
		}

		ct := p.Header.Get("Content-Type")
		if plaintext.MatchString(ct) == true {
			b, err := ioutil.ReadAll(p.Body)
			if err != nil {
				return err
			}
			writer.WriteString("\n")
			writer.Write(b)
			writer.WriteString("\n")
			writer.Flush()
		}

	}

	return nil
}

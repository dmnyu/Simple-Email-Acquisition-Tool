package go_email

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type EmailAccount struct {
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
}

func GetCreds(account string) (EmailAccount, error) {
	credsMap := map[string]EmailAccount{}
	source, err := ioutil.ReadFile("go-email.yml")
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

func GetClient(account EmailAccount) (*client.Client, error) {
	log.Println("Connecting to server...")

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

func GetMessage(i int, c *client.Client) (chan *imap.Message, error) {
	seqset := new(imap.SeqSet)
	seqset.AddNum(uint32(i))
	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	if err := <-done; err != nil {
		return nil, err
	}

	return messages, nil
}

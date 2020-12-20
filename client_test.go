package go_email

import (
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"log"
	"testing"
)

func TestClient(t *testing.T) {
	var account EmailAccount
	var err error
	var imapClient *client.Client
	var mailbox *imap.MailboxStatus
	var name string

	t.Run("Test Parsing Account Config", func(t *testing.T) {
		account, err = GetCreds("don")
		if err != nil {
			t.Error(err)
		}

		t.Log(account)
	})

	t.Run("Test Get Client", func(t *testing.T) {
		imapClient, err = GetClient(account)
		if err != nil {
			t.Error(err)
		}
		t.Log(imapClient)
	})

	t.Run("Test Server Login", func(t *testing.T) {
		err = imapClient.Login(account.Account, account.Password)
		if err != nil {
			t.Error(err)
		}
		t.Log(imapClient)

	})

	t.Run("Test List Mailboxes", func(t *testing.T) {
		mailboxes, err := GetMailboxes(imapClient)
		if err != nil {
			t.Error(err)
		}
		if len(mailboxes) <= 0 {
			t.Error("Received a zero-length slice of mailboxes")
		}
		name = mailboxes[0]
	})

	t.Run("Test Get an Mailbox", func(t *testing.T) {

		mailbox, err = GetMailbox(imapClient, name)
		if err != nil {
			t.Error(err)
		}
		if mailbox.Messages <= 0 {
			t.Errorf("received nil count of messages in mailbox %s", name)
		}

		t.Logf("Flags for %s:%v", name, mailbox.Flags)

	})

	t.Run("Test Get A Message", func(t *testing.T) {
		messages, err := GetMessage(1, imapClient)
		if err != nil {
			t.Error(err)
		}
		for message := range messages {
			log.Println(message.Envelope.Subject)
		}
	})

	t.Run("Test Server Logout", func(t *testing.T) {
		err = imapClient.Logout()
		if err != nil {
			t.Error(err)
		}
	})

}

package go_email

import (
	"flag"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"testing"
)

func TestClient(t *testing.T) {
	var account EmailAccount
	var err error
	var imapClient *client.Client
	var mailbox *imap.MailboxStatus
	var name string
	var email Email

	flag.Parse()

	t.Run("Test Parsing Account Config", func(t *testing.T) {
		account, err = GetCreds("don")
		if err != nil {
			t.Error(err)
		}

		t.Log(account)
	})

	t.Run("Test Get Client", func(t *testing.T) {
		imapClient, err = GetClient("don")
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

		email, err = GetMessage(mailbox.Messages, imapClient)
		if err != nil {
			t.Error(err)
		}

		t.Log(email)
	})

	t.Run("Test print an email message", func(t *testing.T) {
		err := PrintMessage(email)
		if err != nil {
			t.Log(err)
		}
	})

	t.Run("Test Get Messages", func(t *testing.T) {
		mailbox, err = GetMailbox(imapClient, name)
		if err != nil {
			t.Error(err)
		}

		from := uint32(1)
		to := mailbox.Messages
		if mailbox.Messages > 10 {
			from = mailbox.Messages - 10
		}

		messages, err := GetMessages(from, to, imapClient)
		if err != nil {
			t.Error(err)
		}

		for msg := range messages {
			t.Log(msg.SeqNum, msg.Envelope.Subject)
		}

	})

	t.Run("Test Server Logout", func(t *testing.T) {
		err = imapClient.Logout()
		if err != nil {
			t.Error(err)
		}
	})

}

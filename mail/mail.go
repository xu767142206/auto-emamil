package mail

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
)

type MailClient struct {
	config Config
	*client.Client
}

func (mailClient *MailClient) InitImap() (err error) {
	// Connect to server
	mailClient.Client, err = client.DialTLS(fmt.Sprintf("%s:%d", mailClient.config.Addr, mailClient.config.Port), nil)
	if err != nil {
		return
	}
	log.Println("Connected imap")

	// Don't forget to logout
	//defer c.Logout()

	if err = mailClient.Client.Login(mailClient.config.Mail, mailClient.config.Password); err != nil {
		return
	}
	log.Println("Logged in")

	return

}

// NewMail 获取客户端
func NewMail(conf Config) (*MailClient, error) {
	m := new(MailClient)
	m.config = conf
	err := m.InitImap()
	return m, err
}

// GetMessage 获取邮件信息
func (mailClient *MailClient) GetMessage(message *imap.Message, section *imap.BodySectionName) (*mail.Reader, error) {
	r := message.GetBody(section)
	return mail.CreateReader(r)
}

// GetMessageBody 获取邮件信息内容
func (mailClient *MailClient) GetMessageBody(reader *mail.Reader) (body []byte, fileMap map[string][]byte) {
	for {
		p, err := reader.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			break
		}
		if p != nil {
			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				body, err = ioutil.ReadAll(p.Body)
				if err != nil {
					log.Println("read body err:", err.Error())
				}
			case *mail.AttachmentHeader:
				fileName, _ := h.Filename()
				fileData, _ := ioutil.ReadAll(p.Body)
				fileMap[fileName] = fileData
			}
		}
	}
	return
}

// GetMailboxes 同步获取邮箱列表
func (mailClient *MailClient) GetMailboxes() ([]*imap.MailboxInfo, error) {

	boxs := make([]*imap.MailboxInfo, 0)

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- mailClient.List("", "*", mailboxes)
	}()

	err := <-done
	for m := range mailboxes {
		boxs = append(boxs, m)
	}
	return boxs, err

}

// GetMessageList 同步获取邮件列表
func (mailClient *MailClient) GetMessageList(from, to uint32) (imap.BodySectionName, []*imap.Message, error) {

	mails := make([]*imap.Message, 0)

	if from == 0 {
		from = 1
	}

	section, messages, done := mailClient.GetSyncMessageList(from, to)

	err := <-done

	for m := range messages {
		mails = append(mails, m)
	}

	return section, mails, err
}

// GetMessageListLimit 同步获取邮件列表
func (mailClient *MailClient) GetMessageListLimit(limit uint32, mbox *imap.MailboxStatus, orderBy string) (imap.BodySectionName, []*imap.Message, error) {

	from, to := mailClient.GetRange(limit, mbox, orderBy)

	return mailClient.GetMessageList(from, to)
}

// GetSyncMessageListLimit 异步获取邮件列表
func (mailClient *MailClient) GetSyncMessageListLimit(limit uint32, mbox *imap.MailboxStatus, orderBy string) (imap.BodySectionName, <-chan *imap.Message, <-chan error) {

	from, to := mailClient.GetRange(limit, mbox, orderBy)

	return mailClient.GetSyncMessageList(from, to)
}

// GetRange 获取 邮件最近几条 asc 升序 desc 降序
func (mailClient *MailClient) GetRange(limit uint32, mbox *imap.MailboxStatus, orderBy string) (from, to uint32) {

	from = 1
	to = limit

	if orderBy == "desc" {
		if mbox.Messages > limit {
			from = mbox.Messages - limit + 1
		}
		to = mbox.Messages
	}
	return
}

// GetSyncMessageList 异步获取邮件
func (mailClient *MailClient) GetSyncMessageList(from, to uint32) (imap.BodySectionName, <-chan *imap.Message, <-chan error) {

	if from == 0 {
		from = 1
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, to-from+1)
	section := imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	done := make(chan error, 1)
	go func() {
		done <- mailClient.Fetch(seqset, items, messages)
	}()

	return section, messages, done
}

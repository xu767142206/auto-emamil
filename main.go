package main

import (
	"auto-mail/mail"
	"encoding/base64"
	"fmt"
	"github.com/emersion/go-imap"
	"log"
	"strings"
	"time"
)

const INBOX = "INBOX"
const DELBOX = "Deleted Messages"

func DecodeMailHead(str string) string {
	if strings.Contains(str, "?utf-8?B?") {
		replace := strings.Replace(str, "?utf-8?B?", "", 1)
		size := 2
		if len(replace)%2 == 0 {
			size = 4
		}
		base64str := replace[1 : len(replace)-size]
		decodeString, _ := base64.RawStdEncoding.DecodeString(base64str)
		return string(decodeString)
	}
	return str
}

func main() {

	config := mail.NewConfig()
	config.Mail = "1368332201@qq.com"
	config.Password = "beldkofrdjeuifdf"

	mailClient, err := mail.NewMail(config)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		mailClient.Logout()
	}()

	mbox, _ := mailClient.Select(INBOX, false)

	section, messages, err := mailClient.GetMessageListLimit(uint32(10), mbox, "desc")

	if err != nil {
		log.Fatalln(err)
	}
	//删除
	seqset := new(imap.SeqSet)

MESSAGE:
	for _, msg := range messages {
		message, _ := mailClient.GetMessage(msg, &section)
		subject, _ := message.Header.Subject()

		addrs, _ := message.Header.AddressList("From")
		if len(addrs) <= 0 {
			continue
		}
		from := addrs[0]
		date, _ := message.Header.Date()

		fmt.Println(DecodeMailHead(from.Name))
		if from.Address == "10000@qq.com" {
			seqset.AddNum(msg.SeqNum)
			continue
		}

		for _, com := range [...]string{"@tencent.com", "@qt.io", "@alitrade.1688.com"} {
			if strings.Contains(from.Address, com) {
				continue MESSAGE
			}
		}

		for _, emial := range [...]string{"macangshequ@163.com"} {
			if emial == from.Address {
				continue MESSAGE
			}
		}

		// 如果时间超过2017年的
		if date.Unix() > time.Date(2017, time.May, 1, 0, 0, 0, 0, time.Local).Unix() {
			fmt.Println("==============================================================================================")
			fmt.Println("删除邮件")
			fmt.Println("==============================================================================================")
			fmt.Println(DecodeMailHead(subject))
			fmt.Println(from.Name, from.Address)
			seqset.AddNum(msg.SeqNum)
		}
		//body, fileMap := mailClient.GetMessageBody(message)
		//fmt.Println(string(body))
		//fmt.Println(fileMap)
	}

	mailClient.Move(seqset, DELBOX)
}

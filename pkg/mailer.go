package pkg

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"strings"
	"time"
)

var (
	emailMSGBody = `
Hey portfoliotools subscribers,</br>
<br>
Attached is the daily ranges for the list of tickers currently being tracked for %s.</br>
To read this easier, you can pass it through a json parser as the file is in json format.</br>
<br>

Regards,<br>
Bret<br>
<br>
`
)

func SendEmail(smtpCreds StockDataConf, attachmentPath string) (err error) {
	excelPath := strings.Split(attachmentPath, ".")[0] + ".xlsx"
	emailMSGBody = fmt.Sprintf(emailMSGBody, time.Now().UTC().Format(time.RFC3339))
	msgBody := "<html><body>" + emailMSGBody + "</body></html>"
	msg := gomail.NewMessage()
	msg.SetHeader("From", smtpCreds.EmailAddress)
	msg.SetHeader("To", smtpCreds.EmailAddress)
	msg.SetHeader("Bcc", smtpCreds.MailTo...)
	msg.SetHeader("Subject", "Daily Stock VAPR Ranges")
	msg.SetBody("text/html", msgBody)
	msg.Attach(attachmentPath)
	msg.Attach(excelPath)

	d := gomail.NewDialer(smtpCreds.Hostname, smtpCreds.Port, smtpCreds.EmailAddress, smtpCreds.EmailPassword)
	err = d.DialAndSend(msg)
	return err
}

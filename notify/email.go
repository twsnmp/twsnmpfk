// Package notify : 通知処理
package notify

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/wneessen/go-mail"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

func canSendMail() bool {
	switch datastore.NotifyConf.Provider {
	case "google", "microsoft":
		return datastore.HasValidNotifyOAuth2Token(datastore.NotifyConf)
	default:
		if datastore.NotifyConf.MailServer == "" ||
			datastore.NotifyConf.MailFrom == "" ||
			datastore.NotifyConf.MailTo == "" {
			return false
		}
	}
	return true
}

func sendNotifyMail(list []*datastore.EventLogEnt) {
	if !canSendMail() {
		return
	}
	nl := getLevelNum(datastore.NotifyConf.Level)
	if nl == 3 {
		return
	}
	nd := getNotifyData(list, nl)
	if nd.failureBody != "" {
		err := SendMail(nd.failureSubject, nd.failureBody)
		r := ""
		level := "info"
		if err != nil {
			log.Printf("send mail err=%v", err)
			r = fmt.Sprintf("err=%v", err)
			level = "low"
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: level,
			Event: fmt.Sprintf(i18n.Trans("Send notify mail %s"), r),
		})
	}
	if nd.repairBody != "" {
		err := SendMail(nd.repairSubject, nd.repairBody)
		r := ""
		level := "info"
		if err != nil {
			log.Printf("send mail err=%v", err)
			r = fmt.Sprintf("err=%v", err)
			level = "low"
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: level,
			Event: fmt.Sprintf(i18n.Trans("Send repair mail %s"), r),
		})
	}
}

func SendMail(subject, body string) error {
	if !canSendMail() {
		return nil
	}
	switch datastore.NotifyConf.Provider {
	case "google":
		return sendMailOAuth2("smtp.gmail.com", subject, body)
	case "microsoft":
		return sendMailOAuth2("smtp-mail.outlook.com", subject, body)
	default:
		return sendMailSMTP(subject, body)
	}
}

func sendMailSMTP(subject, body string) error {
	host, portStr, err := net.SplitHostPort(datastore.NotifyConf.MailServer)
	if err != nil {
		host = datastore.NotifyConf.MailServer
		portStr = ""
	}

	var options []mail.Option

	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			options = append(options, mail.WithPort(port))
		}
	}

	tlsconfig := &tls.Config{
		ServerName:         host,
		// #nosec G402
		InsecureSkipVerify: datastore.NotifyConf.InsecureSkipVerify,
	}
	if datastore.NotifyConf.InsecureSkipVerify {
		for _, e := range tls.CipherSuites() {
			tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, e.ID)
		}
		tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, tls.TLS_RSA_WITH_AES_128_GCM_SHA256)
		tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, tls.TLS_RSA_WITH_AES_256_GCM_SHA384)
	}
	options = append(options, mail.WithTLSConfig(tlsconfig))

	if strings.HasSuffix(datastore.NotifyConf.MailServer, ":465") {
		options = append(options, mail.WithSSL())
	} else {
		options = append(options, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	if datastore.NotifyConf.User != "" {
		options = append(options,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(datastore.NotifyConf.User),
			mail.WithPassword(datastore.NotifyConf.Password),
		)
	}

	client, err := mail.NewClient(host, options...)
	if err != nil {
		log.Printf("send mail err=%v", err)
		return err
	}

	message := mail.NewMsg()
	if err := message.From(datastore.NotifyConf.MailFrom); err != nil {
		log.Printf("send mail err=%v", err)
		return err
	}
	for _, rcpt := range strings.Split(datastore.NotifyConf.MailTo, ",") {
		if !strings.Contains(rcpt, "@") {
			continue
		}
		if err := message.AddTo(rcpt); err != nil {
			log.Printf("send mail err=%v", err)
			return err
		}
	}

	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, body)

	if err := client.DialAndSend(message); err != nil {
		log.Printf("send mail err=%v", err)
		return err
	}

	log.Printf("send mail to %s", datastore.NotifyConf.MailTo)
	return nil
}

func SendTestMail(testConf *datastore.NotifyConfEnt) error {
	switch testConf.Provider {
	case "google":
		return sendTestMailOAuth2("smtp.gmail.com", testConf)
	case "microsoft":
		return sendTestMailOAuth2("smtp-mail.outlook.com", testConf)
	default:
		return sendTestMailSMTP(testConf)
	}
}

func sendTestMailSMTP(testConf *datastore.NotifyConfEnt) error {
	host, portStr, err := net.SplitHostPort(testConf.MailServer)
	if err != nil {
		host = testConf.MailServer
		portStr = ""
	}

	var options []mail.Option

	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			options = append(options, mail.WithPort(port))
		}
	}

	tlsconfig := &tls.Config{
		ServerName:         host,
		// #nosec G402
		InsecureSkipVerify: testConf.InsecureSkipVerify,
	}
	if testConf.InsecureSkipVerify {
		for _, e := range tls.CipherSuites() {
			tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, e.ID)
		}
		tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, tls.TLS_RSA_WITH_AES_128_GCM_SHA256)
		tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, tls.TLS_RSA_WITH_AES_256_GCM_SHA384)
	}
	options = append(options, mail.WithTLSConfig(tlsconfig))

	if strings.HasSuffix(testConf.MailServer, ":465") {
		options = append(options, mail.WithSSL())
	} else {
		options = append(options, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	if testConf.User != "" {
		options = append(options,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(testConf.User),
			mail.WithPassword(testConf.Password),
		)
	}

	client, err := mail.NewClient(host, options...)
	if err != nil {
		log.Printf("send test mail err=%v", err)
		return err
	}

	message := mail.NewMsg()
	if err := message.From(testConf.MailFrom); err != nil {
		log.Printf("send test mail err=%v", err)
		return err
	}
	for _, rcpt := range strings.Split(testConf.MailTo, ",") {
		if !strings.Contains(rcpt, "@") {
			continue
		}
		if err := message.AddTo(rcpt); err != nil {
			log.Printf("send test mail err=%v", err)
			return err
		}
	}

	t, err := template.New("test").Parse(datastore.LoadMailTemplate("test"))
	if err != nil {
		log.Printf("send test mail err=%s", err)
		return err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, map[string]interface{}{
		"Title": testConf.Subject + i18n.Trans("(test mail)"),
	}); err != nil {
		return err
	}
	body := buffer.String()

	message.Subject(testConf.Subject)
	message.SetBodyString(mail.TypeTextHTML, body)

	if err := client.DialAndSend(message); err != nil {
		log.Printf("send test mail err=%v", err)
		return err
	}

	return nil
}

func sendMailOAuth2(server, subject, body string) error {
	token := getNotifyOAuth2Token()
	if token == nil {
		return fmt.Errorf("oauth2 token not found")
	}
	client, err := mail.NewClient(server,
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthXOAUTH2),
		mail.WithUsername(datastore.NotifyConf.User), mail.WithPassword(token.AccessToken))
	if err != nil {
		return err
	}
	message := mail.NewMsg()
	if err := message.From(datastore.NotifyConf.MailFrom); err != nil {
		return err
	}
	for _, rcpt := range strings.Split(datastore.NotifyConf.MailTo, ",") {
		if !strings.Contains(rcpt, "@") {
			continue
		}
		if err := message.AddTo(rcpt); err != nil {
			return err
		}
	}

	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, body)
	return client.DialAndSend(message)
}

func sendTestMailOAuth2(server string, testConf *datastore.NotifyConfEnt) error {
	token := getNotifyOAuth2Token()
	if token == nil {
		return fmt.Errorf("oauth2 token not found")
	}
	client, err := mail.NewClient(server,
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthXOAUTH2),
		mail.WithUsername(testConf.User), mail.WithPassword(token.AccessToken))
	if err != nil {
		return err
	}
	message := mail.NewMsg()
	if err := message.From(testConf.MailFrom); err != nil {
		return err
	}
	for _, rcpt := range strings.Split(testConf.MailTo, ",") {
		if !strings.Contains(rcpt, "@") {
			continue
		}
		if err := message.AddTo(rcpt); err != nil {
			return err
		}
	}
	t, err := template.New("test").Parse(datastore.LoadMailTemplate("test"))
	if err != nil {
		log.Printf("send test mail err=%s", err)
		return err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, map[string]interface{}{
		"Title": testConf.Subject + i18n.Trans("(test mail)"),
	}); err != nil {
		return err
	}
	body := buffer.String()
	message.Subject(testConf.Subject)
	message.SetBodyString(mail.TypeTextHTML, body)
	return client.DialAndSend(message)
}

package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"verivista/pt/config"
)

var ClientMail *gomail.Dialer

// ConnMailClient 连接邮件服务器
func ConnMailClient() {
	mailConfig := config.Config.Mail

	ClientMail = gomail.NewDialer(mailConfig.Host, mailConfig.Port, mailConfig.User, mailConfig.Pwd)
	ClientMail.TLSConfig = &tls.Config{InsecureSkipVerify: true}
}

// SendAuthMail 发送验证码模板邮件
func SendAuthMail(addr string, username string, code int) error {
	mailConfig := config.Config.Mail
	data := struct {
		Username string
		AuthCode int
	}{
		Username: username,
		AuthCode: code,
	}

	m := gomail.NewMessage()
	m.SetHeader(`From`, mailConfig.User)
	m.SetHeader(`To`, addr)
	m.SetHeader(`Subject`, "Verivista PT验证码")

	tpl, err := template.ParseFiles(fmt.Sprintf("./mail/authTemplate.html"))
	if err != nil {
		return fmt.Errorf("[邮件模版转换失败]: %v", err)
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("[邮件变量填充失败]: %v", err)
	}

	m.SetBody("text/html", buf.String())

	if err := ClientMail.DialAndSend(m); err != nil {
		logrus.Infoln("邮件发送失败: ", err)
	} else {
		logrus.Infoln("邮件发送成功")
	}
	return nil
}

func SendWarnMail(addr string) error {
	mailConfig := config.Config.Mail
	data := struct {
		Email string
	}{
		Email: addr,
	}

	m := gomail.NewMessage()
	m.SetHeader(`From`, mailConfig.User)
	m.SetHeader(`To`, addr)
	m.SetHeader(`Subject`, "Verivista PT提示")

	tpl, err := template.ParseFiles(fmt.Sprintf("./mail/warnTemplate.html"))
	if err != nil {
		return fmt.Errorf("[邮件模版转换失败]: %v", err)
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("[邮件变量填充失败]: %v", err)
	}

	m.SetBody("text/html", buf.String())

	if err := ClientMail.DialAndSend(m); err != nil {
		log.Fatal("邮件发送失败: ", err)
	} else {
		log.Fatal("邮件发送成功")
	}
	return nil
}

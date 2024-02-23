package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
	"html/template"
	"io"
	"net/smtp"
	"os"
	"strconv"
	"verivista/pt/config"
)

// ConnMailClient 连接邮件服务器
func ConnMailClient() (*smtp.Client, error) {
	mailConfig := config.Config.Mail
	smtpServer := mailConfig.Host + ":" + strconv.Itoa(mailConfig.Port)
	clientMail, err := smtp.Dial(smtpServer)
	if err != nil {
		logrus.Errorln("无法连接邮件服务器: ", err)
		return nil, fmt.Errorf("[无法连接邮件服务器]: %v", err)
	}

	tlsConfig := &tls.Config{
		ServerName: "smtp.gmail.com",
	}
	_ = clientMail.StartTLS(tlsConfig)

	auth := smtp.PlainAuth("", mailConfig.User, mailConfig.Pwd, mailConfig.Host)
	if err := clientMail.Auth(auth); err != nil {
		logrus.Errorln("邮件服务器认证失败: ", err)
		return nil, fmt.Errorf("[邮件服务器认证失败]： %v", err)
	}
	return clientMail, nil
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

	mailTempPath := os.Getenv("VERIVISTA_MAIL_PATH")
	if mailTempPath == "" {
		mailTempPath = "./mail"
	}

	clientMail, err := ConnMailClient()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	defer func(clientMail *smtp.Client) {
		err := clientMail.Close()
		if err != nil {
			logrus.Errorln("邮件服务器连接关闭失败：", err)
		}
	}(clientMail)

	// 邮件内容
	tpl, err := template.ParseFiles(fmt.Sprintf(mailTempPath + "/authTemplate.html"))
	if err != nil {
		return fmt.Errorf("[邮件模版转换失败]: %v", err)
	}
	var buf bytes.Buffer
	if err = tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("[邮件变量填充失败]: %v", err)
	}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: Verivista PT验证码\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", addr, buf.Bytes()))
	_ = clientMail.Mail(mailConfig.User)
	_ = clientMail.Rcpt(addr)

	// 发送邮件
	w, _ := clientMail.Data()
	defer func(w io.WriteCloser) {
		err := w.Close()
		if err != nil {
			logrus.Errorln("关闭邮件内容连接失败：", err)
		}
	}(w)
	if _, err := w.Write(msg); err != nil {
		logrus.Errorln("邮件发送失败: ", err)
		return fmt.Errorf("[验证码邮件发送失败]： %v", err)
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

	mailTempPath := os.Getenv("VERIVISTA_MAIL_PATH")
	if mailTempPath == "" {
		mailTempPath = "./mail"
	}

	clientMail, err := ConnMailClient()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	defer func(clientMail *smtp.Client) {
		err := clientMail.Close()
		if err != nil {
			logrus.Errorln("邮件服务器连接关闭失败：", err)
		}
	}(clientMail)

	// 邮件内容
	tpl, err := template.ParseFiles(fmt.Sprintf(mailTempPath + "/warnTemplate.html"))
	if err != nil {
		return fmt.Errorf("[邮件模版转换失败]: %v", err)
	}
	var buf bytes.Buffer
	if err = tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("[邮件变量填充失败]: %v", err)
	}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: Verivista PT提示\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", addr, buf.Bytes()))
	_ = clientMail.Mail(mailConfig.User) // 发件人
	_ = clientMail.Rcpt(addr)            // 收件人

	// 发送邮件
	w, _ := clientMail.Data()
	defer func(w io.WriteCloser) {
		err := w.Close()
		if err != nil {
			logrus.Errorln("关闭邮件内容连接失败：", err)
		}
	}(w)
	if _, err := w.Write(msg); err != nil {
		logrus.Errorln("邮件发送失败: ", err)
		return fmt.Errorf("[被挂提示邮件发送失败]： %v", err)
	} else {
		logrus.Infoln("邮件发送成功")
	}
	return nil
}

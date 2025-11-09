package noti

import (
	"context"
	"fmt"

	"github.com/zrp9/launchl/internal/adapter/log/crane"
	gomail "gopkg.in/mail.v2"
)

type Sender struct {
	endpoint string
	source   string
	user     string
	token    string
	port     int
	logger   crane.Zlogrus
}

func NewSender(endpoint, source, user, token string, port int, logger crane.Zlogrus) Sender {
	return Sender{
		endpoint: endpoint,
		source:   source,
		user:     user,
		token:    token,
		port:     port,
		logger:   logger,
	}
}

func (s Sender) Send(ctx context.Context, to []string, subject, html, txt string) error {
	// url := "https://send.api.mailtrap.io/api/send"
	// method := "POST"
	message := gomail.NewMessage()
	message.SetHeader("From", s.source)
	message.SetHeader("To", to...)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", html)
	message.AddAlternative("text/plain", txt)

	dialer := gomail.NewDialer(s.endpoint, s.port, s.user, s.token)

	if err := dialer.DialAndSend(message); err != nil {
		s.logger.MustError(fmt.Errorf("notification sender failed to send email: %w", err))
		return err
	}
	// payload := strings.NewReader(`{\"from\":{\"email\":\"hello@zrp3.dev\",\"name\":\"Mailtrap Test\"},\"to\":[{\"email\":\"zachpalmer1017@gmail.com\"}],\"subject\":\"You are awesome!\",\"text\":\"Congrats for sending test email with Mailtrap!\",\"category\":\"Integration Test\"}`)

	// client := &http.Client{}
	// req, err := http.NewRequest(method, url, payload)

	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// req.Header.Add("Authorization", "Bearer <YOUR_API_TOKEN>")
	// req.Header.Add("Content-Type", "application/json")

	// res, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// defer res.Body.Close()

	// body, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// fmt.Println(string(body))
	return nil
}

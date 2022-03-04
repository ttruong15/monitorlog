package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ttruong15/monitorLog"
	"github.com/ttruong15/monitorLog/services"

	"github.com/joho/godotenv"
)

func main() {

	env := os.Getenv("ENV")
	envFileName := ".env"

	if env != "" {
		envFileName += "." + env
	}

	envFileLocation := os.Getenv("GOPATH") + "/monitorlog/" + envFileName
	err := godotenv.Load(envFileLocation)
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	log.Println("Start monitor log server")
	ticker := time.NewTicker(60 * time.Second)
	srvClosed := make(chan struct{})
	done := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				log.Println("Run at ", t)
				go processServices()
			}
		}
	}()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs
		log.Println("os kill signal received....")
		close(srvClosed)
	}()

	// blocking call, waiting kill signal
	<-srvClosed

	ticker.Stop()

	// close timer channel
	done <- struct{}{}

	log.Println("exiting!!!")
}

func processServices() {
	var monitorServices []monitorLog.Processor

	for service, _ := range monitorLog.AvailableServices {
		fmt.Println("Service ", service)
		switch service {
		case "fail2ban":
			monitorServices = append(monitorServices, services.NewFail2banService())
		}
	}

	from := os.Getenv("ML_MAIL_FROM")
	username := os.Getenv("ML_MAIL_USERNAME")
	password := os.Getenv("ML_MAIL_PASSWORD")
	to := []string{os.Getenv("ML_MAIL_TO")}
	smtpHost := os.Getenv("ML_MAIL_HOST")
	smtpPort := os.Getenv("ML_MAIL_PORT")

	auth := smtp.PlainAuth("", username, password, smtpHost)

	header := make(map[string]string)
	header["From"] = from
	header["To"] = os.Getenv("ML_MAIL_TO")
	header["Subject"] = "fail2ban"
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	for _, monitorService := range monitorServices {
		err := monitorService.Processing()
		if err != nil {
			log.Println(err)
			continue
		}
		report := monitorService.GenerateReport()

		message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(report))
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(message))
		if err != nil {
			log.Println(err)
		}
	}
}

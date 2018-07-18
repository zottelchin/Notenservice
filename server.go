package main

import (
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/spf13/viper"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"

	"github.com/gin-gonic/gin"

	"./ovgunoten"
)

var saveState []ovgunoten.Klausur

func main() {
	viper.SetConfigFile("config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
	}

	if viper.GetBool("webpage") {
		go routine()
		fmt.Println("Starting Webserver...")
		gin.SetMode(gin.ReleaseMode)
		router := gin.Default()
		router.StaticFile("/", "frontend/notenuebersicht.html")
		router.StaticFile("/milligram.min.css", "frontend/milligram.min.css")
		router.StaticFile("/vue.min.js", "frontend/vue.min.js")
		//router.GET("/von/:account/:password", func(c *gin.Context) {
		//	c.JSON(200, ovgunoten.InsertToDB(c.Param("account"), c.Param("password")))
		//})
		router.GET("/von/me", func(c *gin.Context) {
			c.JSON(200, saveState)
		})
		router.Run(":3412")
	} else {
		routine()
	}
}

func sendMessage(neu []ovgunoten.Klausur) (string, error) {
	mg := mailgun.NewMailgun(viper.GetString("mailgun.domain"), viper.GetString("mailgun.api-key"), "")
	m := mg.NewMessage(
		viper.GetString("mailgun.sender-name")+" <no-reply@"+viper.GetString("mailgun.domain")+">",
		viper.GetString("mailgun.subject"),
		"Hey, \n\n Im LSF ist eine neue Note aufgetaucht. Folgende Noten sind aufgetaucht: \n\n"+
			ovgunoten.NotenAlsString(neu)+
			"\n Gehe auf "+viper.GetString("domain")+" um alle deine Noten in der Übersicht zu sehen.",
		viper.GetString("mailgun.reciver"),
	)
	_, id, err := mg.Send(m)
	log.Println("Mailnotification sent to " + viper.GetString("mailgun.reciver"))
	return id, err
}

func send(neu []ovgunoten.Klausur) {

	msg := "From: " + viper.GetString("mail.sender") + "\n" +
		"To: " + viper.GetString("mail.reciver") + "\n" +
		"Subject: Neue Note! " + "\n\n" + "Hey, \n\n Im LSF ist eine neue Note aufgetaucht. Folgende Noten sind aufgetaucht: \n\n" +
		ovgunoten.NotenAlsString(neu) +
		"\n Gehe auf " + viper.GetString("domain") + " um alle deine Noten in der Übersicht zu sehen."

	err := smtp.SendMail(viper.GetString("mail.smtpserver")+":"+viper.GetString("mail.smtpport"),
		smtp.PlainAuth("", viper.GetString("mail.sender"), viper.GetString("mail.password"), viper.GetString("mail.smtpserver")),
		viper.GetString("mail.sender"), viper.GetStringSlice("mail.reciver"), []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Println("Mailnotification sent to " + viper.GetString("mail.reciver"))
}

func routine() {
	fmt.Println("Starting Routine...")
	tmp := ovgunoten.InsertToDB(viper.GetString("lsf.user"), viper.GetString("lsf.password"))
	fmt.Println("Got Grades")
	diff := difference(saveState, tmp)
	if len(diff) > 0 {
		if viper.GetBool("smtpmail-mail") {
			send(diff)
		}
		if viper.GetBool("mailgun-mail") {
			sendMessage(diff)
		}
		saveState = tmp
	}
	time.Sleep(time.Hour)
	routine()
}

func difference(alt []ovgunoten.Klausur, neu []ovgunoten.Klausur) []ovgunoten.Klausur {
	result := []ovgunoten.Klausur{}
	for i := 0; i < len(neu); i++ {
		found := false
		for j := 0; j < len(alt); j++ {
			if alt[j].Name == neu[i].Name {
				found = true
			}
		}
		if !found {
			result = append(result, neu[i])
		}
	}
	return result
}

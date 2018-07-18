package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"

	"github.com/gin-gonic/gin"

	"./ovgunoten"
)

var saveState []ovgunoten.Klausur

func main() {
	viper.SetConfigFile("/var/noten/config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
	}

	if viper.GetBool("webpage") {
		go routine()
		fmt.Println("Starting Webserver...")
		gin.SetMode(gin.ReleaseMode)
		router := gin.Default()
		router.StaticFile("/", "/var/noten/frontend/notenuebersicht.html")
		router.StaticFile("/milligram.min.css", "/var/noten/frontend/milligram.min.css")
		router.StaticFile("/vue.min.js", "/var/noten/frontend/vue.min.js")
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
			"\n Gehe auf https://localhost:3412 um alle deine Noten in der Übersicht zu sehen.",
		viper.GetString("mailgun.reciver"),
	)
	_, id, err := mg.Send(m)
	log.Println("Mailnotification sent to " + viper.GetString("mailgun.reciver"))
	return id, err
}

func routine() {
	fmt.Println("Starting Routine...")
	tmp := ovgunoten.InsertToDB(viper.GetString("lsf.user"), viper.GetString("lsf.password"))
	fmt.Println("Got Grades")
	diff := difference(saveState, tmp)
	if len(diff) > 0 {
		sendMessage(diff)
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

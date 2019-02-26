package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"

	"github.com/spf13/viper"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"

	"github.com/zottelchin/Notenservice/ovgunoten"
)

func main() {
	viper.SetConfigFile("config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err.Error())
	}
	routine()
}

func send(neu []ovgunoten.Klausur) {
	if viper.GetBool("smtpmail-mail") {
		msg := "From: " + viper.GetString("mail.sender") + "\n" +
			"To: " + viper.GetString("mail.reciver") + "\n" +
			"Content-Type: text/plain; charset=\"utf-8\"\n" +
			"Subject: Neue Note! " + "\n\n" + "Hey, \n\n Im LSF ist eine neue Note aufgetaucht. Folgende Noten sind aufgetaucht: \n\n" +
			ovgunoten.NotenAlsString(neu) +
			"\n Gehe auf " + viper.GetString("domain") + " um alle deine Noten in der Übersicht zu sehen."

		viper.SetDefault("mail.username", viper.GetString("mail.sender"))
		err := smtp.SendMail(viper.GetString("mail.smtpserver")+":"+viper.GetString("mail.smtpport"),
			smtp.PlainAuth("", viper.GetString("mail.username"), viper.GetString("mail.password"), viper.GetString("mail.smtpserver")),
			viper.GetString("mail.sender"), viper.GetStringSlice("mail.reciver"), []byte(msg))

		if err != nil {
			log.Printf("smtp error: %s", err)
			return
		}

		log.Println("Mailnotification sent to " + viper.GetString("mail.reciver"))
	}
	if viper.GetBool("mailgun-mail") {
		mg := mailgun.NewMailgun(viper.GetString("mailgun.domain"), viper.GetString("mailgun.api-key"), "")
		m := mg.NewMessage(
			viper.GetString("mailgun.sender-name")+" <no-reply@"+viper.GetString("mailgun.domain")+">",
			viper.GetString("mailgun.subject"),
			"Hey, \n\n Im LSF ist eine neue Note aufgetaucht. Folgende Noten sind aufgetaucht: \n\n"+
				ovgunoten.NotenAlsString(neu)+
				"\n Gehe auf "+viper.GetString("domain")+" um alle deine Noten in der Übersicht zu sehen.",
			viper.GetString("mailgun.reciver"),
		)
		_, _, err := mg.Send(m)
		if err != nil {
			log.Printf("Mailgun error: %s", err)
			return
		}
		log.Println("Mailnotification sent to " + viper.GetString("mailgun.reciver"))
	}
}

func routine() {

	log.Println("Starting Routine...")
	tmp := ovgunoten.NotenAbrufen(viper.GetString("lsf.user"), viper.GetString("lsf.password"))

	if len(tmp) == 0 {
		senderr("Leeres Array aus package ovgunoten.")
	} else {
		log.Println("Got Grades")
		log.Printf("Antwort: %s\n", ovgunoten.NotenAlsString(tmp))
	}

	diff := difference(get(), tmp)
	log.Printf("Differenz: %s\n", ovgunoten.NotenAlsString(diff))

	if len(diff) > 0 {
		send(diff)
		save(diff)
	}

}

func difference(alt []ovgunoten.Klausur, neu []ovgunoten.Klausur) []ovgunoten.Klausur {
	result := []ovgunoten.Klausur{}
	for i := 0; i < len(neu); i++ {
		found := false
		for j := 0; j < len(alt); j++ {
			if alt[j].Name == neu[i].Name && alt[j].CP == neu[i].CP && alt[j].Note == neu[i].Note && alt[j].Prüfungszeitraum == neu[i].Prüfungszeitraum {
				found = true
			}
		}
		if !found {
			result = append(result, neu[i])
		}
	}
	return result
}

func senderr(err string) {
	log.Printf("Achtung: %s", err)
	if viper.GetBool("smtpmail-mail") {
		msg := "From: " + viper.GetString("mail.sender") + "\n" +
			"To: " + viper.GetString("mail.reciver") + "\n" +
			"Subject: Fehlermeldung Notenservice " + "\n\n" + "Hey, \n\n Im Notenservice ist ein Fehler aufgetreten: \n\n" +
			err +
			"\n\n Gehe auf " + viper.GetString("domain") + " um alle deine Noten in der Übersicht zu sehen."

		err := smtp.SendMail(viper.GetString("mail.smtpserver")+":"+viper.GetString("mail.smtpport"),
			smtp.PlainAuth("", viper.GetString("mail.sender"), viper.GetString("mail.password"), viper.GetString("mail.smtpserver")),
			viper.GetString("mail.sender"), viper.GetStringSlice("mail.reciver"), []byte(msg))

		if err != nil {
			log.Printf("smtp error: %s\n", err)
			return
		}

		log.Printf("Errormail sent to %s\n", viper.GetString("mail.reciver"))
	}
	if viper.GetBool("mailgun-mail") {
		mg := mailgun.NewMailgun(viper.GetString("mailgun.domain"), viper.GetString("mailgun.api-key"), "")
		m := mg.NewMessage(
			viper.GetString("mailgun.sender-name")+" <no-reply@"+viper.GetString("mailgun.domain")+">",
			"Fehlermeldung Notenservice!",
			"Hey, \n\n Im Notenservie ist ein Fehler aufgetreten: \n\n"+
				err+
				"\n\n Gehe auf "+viper.GetString("domain")+" um alle deine Noten in der Übersicht zu sehen.",
			viper.GetString("mailgun.reciver"),
		)
		_, _, err := mg.Send(m)
		if err != nil {
			log.Printf("mailgun error: %s\n", err)
			return
		}
		log.Printf("Errormail sent to %s\n", viper.GetString("mailgun.reciver"))
	}
}

func save(s []ovgunoten.Klausur) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		log.Printf("Save failed! %s", err)
		return
	}
	err = ioutil.WriteFile(viper.GetString("storage"), jsonData, 0644)
	if err != nil {
		log.Printf("Save failed! %s", err)
		return
	}
	log.Printf("In der Datei gespeichert: %s\n", ovgunoten.NotenAlsString(s))
}

func get() []ovgunoten.Klausur {
	jsonFile, err := os.Open(viper.GetString("storage"))
	if err != nil {
		log.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	res := []ovgunoten.Klausur{}
	json.Unmarshal(byteValue, &res)
	log.Printf("Aus der Datei: %s\n", ovgunoten.NotenAlsString(res))
	return res
}

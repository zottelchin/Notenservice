package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/zottelchin/Notenservice/ovgunoten"
)

func main() {
	viper.SetConfigFile("config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err.Error())
	}
	if viper.GetBool("webpage") {
		log.Println("Starting Webserver...")
		gin.SetMode(gin.ReleaseMode)
		router := gin.Default()
		router.StaticFile("/", "frontend/notenuebersicht.html")
		router.StaticFile("/milligram.min.css", "frontend/milligram.min.css")
		router.StaticFile("/vue.min.js", "frontend/vue.min.js")
		router.StaticFile("/favicon.ico", "frontend/favicon.ico")
		router.StaticFile("/favicon.png", "frontend/favicon.png")
		router.GET("/von/me", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"noten": get(),
			})
		})
		router.Run(":3412")
	}
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

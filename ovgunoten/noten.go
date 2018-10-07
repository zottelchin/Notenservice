package ovgunoten

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var superReturn []Klausur
var count int
var err error

type Klausur struct {
	Name             string `json:"Name"`
	Prüfungszeitraum string `json:"Zeitraum"`
	Note             string `json:"Note"`
	Bestanden        string `json:"Bestanden"`
	CP               string `json:"CP"`
}

func NotenAlsString(noten []Klausur) string {
	result := ""
	for i := 0; i < len(noten); i++ {
		result += "\t[" + noten[i].Bestanden + "] " + noten[i].Name + " (" + noten[i].Prüfungszeitraum + "): " + noten[i].Note + "\n"
	}
	return result
}

func NotenAbrufen(us string, pw string) []Klausur {
	//Don't crash if panic acours
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	//Create Client
	httpClient, CookieJar := clientErstellen(us)

	//Finalen Cookie für den Login abrufen
	DERCookie, err := login(us, pw, httpClient, CookieJar)
	if err != nil {
		log.Println("Ein Problem beim Login")
	}

	asi, err := asiGetter(DERCookie, httpClient)
	if err != nil {
		log.Printf("There was an Error getting the ASI key for %s \nThe error was: %s \n", us, err)
	}

	NotenParsen(DERCookie, asi, httpClient)

	return superReturn
}

func clientErstellen(us string) (httpClient *http.Client, CookieJar *cookiejar.Jar) {
	//Create a new Cookiejar
	CookieJar, err := cookiejar.New(nil)
	if err != nil {
		log.Printf("There was an Error creating the Cookiejar for %s \nThe error was: %s \n", us, err)
	}

	//Create HTTP CLient
	httpClient = &http.Client{
		Jar:     CookieJar,
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	return
}

func login(us string, pw string, client *http.Client, CookieJar *cookiejar.Jar) (*http.Cookie, error) {
	LSFRedirURL := "https://lsf.ovgu.de/Shibboleth.sso/Login?target=/qislsf/rds?state=user&type=1"
	LoginURL := "https://idp-serv.uni-magdeburg.de/idp/Authn/UserPassword?j_password=" + pw + "&j_username=" + us
	GetFinalCookieURL := "https://lsf.ovgu.de/Shibboleth.sso/SAML2/POST"
	SSOFirstCookieUrl := "https://idp-serv.uni-magdeburg.de:443/idp/Authn/UserPassword"
	SSOSecondCookieURL := "https://idp-serv.uni-magdeburg.de:443/idp/profile/SAML2/Redirect/SSO"

	//get first Cookie
	nextURL := LSFRedirURL
	for i := 0; i < 10; i++ {
		resp, _ := client.Get(nextURL)
		if resp.StatusCode == 200 {
			break
		} else {
			nextURL = resp.Header.Get("Location")
		}
	}

	//safe first cookie
	url1, err := url.Parse(SSOFirstCookieUrl)
	if err != nil {
		return nil, err
	}
	firstCookie := CookieJar.Cookies(url1)[0]
	log.Println("First Cookie for " + us)

	//getting second cookie and params
	var resp *http.Response
	nextURL = LoginURL
	for i := 0; i < 10; i++ {
		resp, _ = client.Post(nextURL, "", nil)
		if resp.StatusCode == 200 {
			break
		} else {
			nextURL = resp.Header.Get("Location")
		}
	}

	//second cookie
	url2, err := url.Parse(SSOSecondCookieURL)
	if err != nil {
		return nil, err
	}
	secondCookie := CookieJar.Cookies(url2)[0]
	fmt.Println("Second Cookie for " + us)

	//params
	defer resp.Body.Close()
	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	data := html.UnescapeString(string(c))
	getvalue := regexp.MustCompile("value=\".*\"")
	values := getvalue.FindAllStringSubmatch(data, -1)
	values[0][0] = strings.TrimSuffix(values[0][0], "\"")
	values[0][0] = strings.TrimPrefix(values[0][0], "value=\"")
	values[1][0] = strings.TrimSuffix(values[1][0], "\"")
	values[1][0] = strings.TrimPrefix(values[1][0], "value=\"")

	v := url.Values{
		"SAMLResponse": {values[1][0]},
		"RelayState":   {values[0][0]},
	}

	body := strings.NewReader(v.Encode())
	fmt.Println("Values for " + us)

	//adding values and cookies to request
	req, err := http.NewRequest("POST", GetFinalCookieURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(firstCookie)
	req.AddCookie(secondCookie)
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	//we got the real cookie
	url3, err := url.Parse("https://lsf.ovgu.de/qislsf")
	if err != nil {
		return nil, err
	}
	finalCookie := CookieJar.Cookies(url3)[0]
	fmt.Println("final Cookie :) for " + us)

	return finalCookie, nil
}

func asiGetter(finalCookie *http.Cookie, httpClient *http.Client) (string, error) {
	AsiURL := "https://lsf.ovgu.de/qislsf/rds?state=user&type=1"
	LinkPrüfungsverwaltung := "https://lsf.ovgu.de/qislsf/rds?state=change&type=1&moduleParameter=studyPOSMenu&nextdir=change&next=menu.vm&subdir=applications&xml=menu&purge=y&navigationPosition=functions%2CstudyPOSMenu&breadcrumb=studyPOSMenu&topitem=functions&subitem=studyPOSMenu"

	req, err := http.NewRequest("GET", AsiURL, nil)
	if err != nil {
		return "", err
	}
	req.AddCookie(finalCookie)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	//get asi
	req, err = http.NewRequest("GET", LinkPrüfungsverwaltung, nil)
	if err != nil {
		return "", err
	}
	req.AddCookie(finalCookie)
	resp, err = httpClient.Do(req)
	if err != nil {
		return "", err
	}
	reg := regexp.MustCompile("asi=(.+?)\"")
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	asis := reg.FindAllString(string(data), -1)
	asis[0] = strings.TrimSuffix(asis[0], "\"")
	fmt.Println("Asi :)")
	return asis[0], nil
}

func NotenParsen(finalCookie *http.Cookie, asi string, httpClient *http.Client) {
	URLtoTable := "https://lsf.ovgu.de/qislsf/rds?state=notenspiegelStudent&next=list.vm&nextdir=qispos/notenspiegel/student&createInfos=Y&struct=auswahlBaum&nodeID=auswahlBaum%7Cabschluss%3Aabschl%3D82%2Cstgnr%3D1%2CdeAbschlTxt%3DBachelor+of+Science&expand=0&" + asi + "#auswahlBaum%7Cabschluss%3Aabschl%3D82%2Cstgnr%3D1%2CdeAbschlTxt%3DBachelor+of+Science"

	req, err := http.NewRequest("GET", URLtoTable, nil)
	if err != nil {
		panic(err)
	}
	req.AddCookie(finalCookie)
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	traverse(doc)
}

func traverse(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "table" {
			if count == 1 {
				getTableToDB(c)
			}
			count++
		} else {
			traverse(c)
		}
	}
}

func getTableToDB(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tbody" {
			parseTbody(c)
		}
	}
}

func parseTbody(n *html.Node) {
	allRows := []Klausur{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tr" {
			var row []string
			for x := c.FirstChild; x != nil; x = x.NextSibling {
				for y := x.FirstChild; y != nil; y = y.NextSibling {
					row = append(row, strings.TrimSpace(y.Data))
				}

			}
			if row[1] != "b" && len(row) > 8 {
				result := Klausur{}
				result.Name = row[1]
				result.Prüfungszeitraum = row[2]
				result.Note = row[3]
				if len(row) == 9 {
					result.CP = strings.Trim(row[5], ",0")
					result.Bestanden = row[4]
				} else {
					result.CP = strings.Trim(row[7], ",0")
					result.Bestanden = row[6]
				}
				//fmt.Println()
				//fmt.Println(row)
				//fmt.Println(len(row))
				//fmt.Println(result.Name + " (" + result.Prüfungszeitraum + ")[" + result.CP + " CP]: " + result.Note + "; [" + result.Bestanden + "]")
				allRows = append(allRows, result)

			}

		}
	}
	superReturn = allRows

}

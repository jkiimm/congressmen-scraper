package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Congressman struct {
	Name       string   `json:"name"`
	Image      string   `json:"image"`
	Party      string   `json:"party"`
	District   string   `json:"district"`
	Committees []string `json:"committees"`
	Elections  []int    `json:"elections"`
}

func GetManIds() []string {
	doc := FetchBaseQueryDoc()
	r, _ := regexp.Compile("[0-9]+")

	return doc.Find(".memberna_list .img a").Map(func(i int, s *goquery.Selection) string {
		text, _ := s.Attr("onclick")
		return r.FindString(text)
	})
}

func FetchBaseQueryDoc() *goquery.Document {
	urlstr := "http://www.assembly.go.kr/assm/memact/congressman/memCond/memCondListAjax.do"

	res, err := http.PostForm(urlstr, url.Values{
		"currentPage": {"1"},
		"rowPerPage":  {"300"},
	})

	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromResponse(res)

	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func FetchDetailQueryDoc(id string) *goquery.Document {
	urlpre := "http://www.assembly.go.kr/assm/memPop/memPopup.do?dept_cd="
	doc, _ := goquery.NewDocument(urlpre + id)
	return doc
}

func FillUpManData(id string) *Congressman {
	pre := "http://www.assembly.go.kr"
	r, _ := regexp.Compile("[0-9]+")

	doc := FetchDetailQueryDoc(id)
	detail := doc.Find(".info_mna")
	profile := detail.Find(".profile")
	sub := detail.Find(".pro_detail")

	imagePath, _ := profile.Find(".photo img").Attr("src")
	committees := []string{}
	for _, cmt := range strings.Split(sub.Find("dd:nth-of-type(3)").Text(), ",") {
		committees = append(committees, strings.TrimSpace(cmt))
	}

	elections := []int{}
	for _, numstr := range r.FindAllString(sub.Find("dd:nth-of-type(4)").Text(), -1) {
		num, _ := strconv.Atoi(numstr)
		elections = append(elections, num)
	}

	man := Congressman{
		Name:       profile.Find("h4").Text(),
		Image:      pre + imagePath,
		Party:      sub.Find("dd:nth-of-type(1)").Text(),
		District:   sub.Find("dd:nth-of-type(2)").Text(),
		Committees: committees,
		Elections:  elections,
	}

	return &man
}

func SaveJSON(men []*Congressman) {
	json, _ := json.MarshalIndent(men, "", "  ")
	fmt.Printf("%s", string(json))

	err := ioutil.WriteFile("output.json", json, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	ids := GetManIds()

	men := []*Congressman{}
	for i, id := range ids {
		man := FillUpManData(id)

		men = append(men, man)
		fmt.Printf("%s %d/%d\n", men[len(men)-1].Name, i, len(ids))
	}

	SaveJSON(men)
}

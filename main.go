package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type jobs struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

var baseURL string = "https://kr.indeed.com/취업?as_and=python&radius=25&l=seoul&limit=50"
var classCard string = ".jobsearch-SerpJobCard"
var classJobID string = ".data-jk"

func main() {
	var allJobs []jobs
	totalPages := getNumberOfPages()
	for i := 0; i < totalPages; i++ {
		jobsInPage := getPage(i)
		allJobs = append(allJobs, jobsInPage...)
	}
	fmt.Println(allJobs)
}

func getPage(page int) []jobs {
	var jobsInPage []jobs
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting: ", pageURL)
	res, err := http.Get(pageURL)
	checkError(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	searchCards := doc.Find(classCard)

	searchCards.Each(func(i int, card *goquery.Selection) {
		job := extractJob(card)
		jobsInPage = append(jobsInPage, job)
	})
	return jobsInPage
}

func extractJob(card *goquery.Selection) jobs {
	id, _ := card.Attr(classJobID)
	title := card.Find(".title>a").Text()
	location := card.Find(".sjcl").Text()
	salary := card.Find(".salaryText").Text()
	summary := card.Find(".summary").Text()
	return jobs{
		id:       cleanString(id),
		title:    cleanString(title),
		location: cleanString(location),
		salary:   cleanString(salary),
		summary:  cleanString(summary),
	}
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getNumberOfPages() int {
	pages := 0
	fmt.Println("Getting the number of pages...")
	res, err := http.Get(baseURL)
	checkError(err)
	checkStatusCode(res)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})
	return pages
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatusCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status: ", res.StatusCode)
	}
}

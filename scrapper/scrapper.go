package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type job struct {
	id       string
	title    string
	company  string
	location string
	salary   string
	summary  string
}

// Scrape : scrape jobs from indeed.com
func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/취업?as_and=" + term + "&radius=25&l=seoul&limit=50"
	startTime := time.Now()
	var allJobs []job
	c := make(chan []job)
	totalPages := getNumberOfPages(baseURL)
	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, c)
	}
	for i := 0; i < totalPages; i++ {
		jobsInPage := <-c
		allJobs = append(allJobs, jobsInPage...)
	}
	writeJobs(allJobs)
	fmt.Println("Done, extract", len(allJobs), "jobs from indeed.com")
	endTime := time.Now()
	fmt.Println("Operating time: ", endTime.Sub(startTime))
}

func writeJobs(allJobs []job) {
	c := make(chan []string)
	file, err := os.Create("Indeed_Jobs.csv")
	checkError(err)
	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Title", "Company", "Location", "Salary", "Link", "Summary"}
	writeErr := w.Write(headers)
	checkError(writeErr)

	for _, job := range allJobs {
		go writeJobDetail(job, c)
	}

	for i := 0; i < len(allJobs); i++ {
		jobData := <-c
		writeErr := w.Write(jobData)
		checkError(writeErr)
	}
}

func writeJobDetail(job job, c chan<- []string) {
	const jobURL = "https://kr.indeed.com/viewjob?jk="
	c <- []string{job.title, job.company, job.location, job.salary, jobURL + job.id, job.summary}
}

func getPage(page int, url string, upperC chan<- []job) {
	const classCard = ".jobsearch-SerpJobCard"
	const pageConnection = "&start="
	var jobsInPage []job
	c := make(chan job)
	pageURL := url + pageConnection + strconv.Itoa(page*50)
	fmt.Println("Requesting: ", pageURL)
	res, err := http.Get(pageURL)
	checkError(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	searchCards := doc.Find(classCard)

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})
	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobsInPage = append(jobsInPage, job)
	}
	upperC <- jobsInPage
}

func extractJob(card *goquery.Selection, c chan<- job) {
	const classTitle = ".title>a"
	const classCompany = ".company"
	const classLocation = ".location"
	const classSalary = ".salaryText"
	const classSummary = ".summary"
	const classJobID = ".data-jk"
	id, _ := card.Attr(classJobID)
	title := card.Find(classTitle).Text()
	company := card.Find(classCompany).Text()
	location := card.Find(classLocation).Text()
	salary := card.Find(classSalary).Text()
	summary := card.Find(classSummary).Text()
	c <- job{
		id:       CleanString(id),
		title:    CleanString(title),
		company:  CleanString(company),
		location: CleanString(location),
		salary:   CleanString(salary),
		summary:  CleanString(summary),
	}
}

// CleanString : trim space and join words
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getNumberOfPages(url string) int {
	const classPagination = ".pagination"
	pages := 0
	fmt.Println("Getting the number of pages...")
	res, err := http.Get(url)
	checkError(err)
	checkStatusCode(res)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)
	doc.Find(classPagination).Each(func(i int, s *goquery.Selection) {
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

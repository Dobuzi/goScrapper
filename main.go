package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

const baseURL = "https://kr.indeed.com/취업?as_and=python&radius=25&l=seoul&limit=50"
const jobURL = "https://kr.indeed.com/viewjob?jk="
const pageConnection = "&start="
const classPagination = ".pagination"
const classTitle = ".title>a"
const classCompany = ".company"
const classLocation = ".location"
const classSalary = ".salaryText"
const classSummary = ".summary"
const classCard = ".jobsearch-SerpJobCard"
const classJobID = ".data-jk"

func main() {
	var allJobs []job
	totalPages := getNumberOfPages()
	for i := 0; i < totalPages; i++ {
		jobsInPage := getPage(i)
		allJobs = append(allJobs, jobsInPage...)
	}
	writeJobs(allJobs)
	fmt.Println("Done, extract", len(allJobs), "jobs from indeed.com")
}

func writeJobs(allJobs []job) {
	file, err := os.Create("Indeed_Jobs.csv")
	checkError(err)
	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Title", "Company", "Location", "Salary", "Link", "Summary"}
	writeErr := w.Write(headers)
	checkError(writeErr)

	for _, job := range allJobs {
		jobData := []string{job.title, job.company, job.location, job.salary, jobURL + job.id, job.summary}
		writeErr := w.Write(jobData)
		checkError(writeErr)
	}
}

func getPage(page int) []job {
	var jobsInPage []job
	pageURL := baseURL + pageConnection + strconv.Itoa(page*50)
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

func extractJob(card *goquery.Selection) job {
	id, _ := card.Attr(classJobID)
	title := card.Find(classTitle).Text()
	company := card.Find(classCompany).Text()
	location := card.Find(classLocation).Text()
	salary := card.Find(classSalary).Text()
	summary := card.Find(classSummary).Text()
	return job{
		id:       cleanString(id),
		title:    cleanString(title),
		company:  cleanString(company),
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

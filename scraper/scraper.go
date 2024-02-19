package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/jordan-wright/email"
	"log"
	"net/smtp"
	"strings"
	"time"
)

type Email struct {
	SenderEmail string // account email
	Password    string // account password
	SmtpHost    string // host
	SmtpPort    string // host port
}

func NewEmail(senderEmail string, password string, smtpHost string, smtpPort string) *Email {
	newEmail := &Email{
		SenderEmail: senderEmail,
		Password:    password,
		SmtpHost:    smtpHost,
		SmtpPort:    smtpPort,
	}
	return newEmail
}

func (e *Email) SendEmail(recipient string, url string) {
	newEmail := email.NewEmail()

	newEmail.From = e.SenderEmail
	newEmail.To = []string{recipient}
	newEmail.Subject = "Web Scraper Alert: Item in Stock!"
	body := fmt.Sprintf("Item is in stock at: %s\n\nThis email was generated automatically.", url)
	newEmail.Text = []byte(body)

	err := newEmail.Send(fmt.Sprintf("%s:%s", e.SmtpHost, e.SmtpPort), smtp.PlainAuth("", e.SenderEmail, e.Password, e.SmtpHost))
	if err != nil {
		fmt.Println("Mail failed to send")
		fmt.Println(err)
	}
}

type Job struct {
	ClassEncounter string // html class to scan
	Element        string // element to retrieve value on
	PositiveValue  string // expected positive value
	NegativeValue  string // expected negative value
}

type Scraper struct {
	Url            string // url to scrape
	RecipientEmail string // email to alert on positive trigger
	Jobs           []*Job // slice of objects to scrape for
	WaitTime       int    // time to wait between request (minutes)
	Email          *Email
}

func NewScraper(url string, recipientEmail string, waitTime int, email *Email) *Scraper {
	newScraper := &Scraper{
		Url:            url,
		RecipientEmail: recipientEmail,
		Jobs:           []*Job{},
		WaitTime:       waitTime,
		Email:          email,
	}
	return newScraper
}

func (s *Scraper) AddJob(classEncounter string, element string, positiveValue string, negativeValue string) {
	job := &Job{
		ClassEncounter: classEncounter,
		Element:        element,
		PositiveValue:  positiveValue,
		NegativeValue:  negativeValue,
	}
	s.Jobs = append(s.Jobs, job)
}

func (s *Scraper) addJobs(c *colly.Collector) {
	for _, job := range s.Jobs {
		// Set up callbacks to handle scraping events
		c.OnHTML(job.ClassEncounter, func(e *colly.HTMLElement) {
			// Extract data from HTML elements
			data := e.ChildText(job.Element)
			// Clean up the extracted data
			data = strings.TrimSpace(data)

			// Print the scraped data
			fmt.Printf("Data: %s\n", data)
			if data == job.PositiveValue {
				fmt.Println("Website " + s.Url + ": positive value detected - " + job.PositiveValue + " - " + time.Now().String())
				s.Email.SendEmail(s.RecipientEmail, s.Url)
			} else if data == job.NegativeValue {
				fmt.Println("Website " + s.Url + ": negative value detected - " + job.NegativeValue + " - " + time.Now().String())
			} else {
				fmt.Println("Website " + s.Url + ": unknown response - " + time.Now().String())
			}
		})
	}
}

func (s *Scraper) Scrape() {
	// Create a new Colly collector
	c := colly.NewCollector()
	s.addJobs(c)

	// Create a ticker that ticks once per hour
	ticker := time.NewTicker(time.Duration(s.WaitTime) * time.Hour)

	for {
		select {
		case <-ticker.C: // This block will be executed once every hour
			// Visit the URL and start scraping
			err := c.Visit(s.Url)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

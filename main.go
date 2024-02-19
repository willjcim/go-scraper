package main

import (
	"github.com/willjcim/scraper/scraper"
)

func main() {
	email := scraper.NewEmail("", "", "smtp.gmail.com", "587")
	s := scraper.NewScraper("", "", 15, email)
	s.AddJob("", "", "", "")
	s.Scrape()
}

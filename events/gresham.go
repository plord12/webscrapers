/**

find gresham events

*/

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
	"github.com/playwright-community/playwright-go"
)

func gresham() {

	ebPage := 1

	// loop through all pages until we get nothing more ... store results in array for later sorting
	//
	url := "https://www.gresham.ac.uk/whats-on?see-all"

	fmt.Fprintf(os.Stderr, "Fetching %s\n", url)
	_, err := page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		_, err = page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not goto url: %v\n", err)
			return
		}
	}

	// reject cookie
	//
	page1.GetByText("Customize", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})
	page1.GetByText("Deny", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

	for {
		if ebPage > cliOptions.Maxpage {
			break
		}

		events, err := page1.Locator(".o-teaser__content").Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).All()
		if err != nil || len(events) == 0 {
			// no more pages
			break
		}

		for _, event := range events {
			eventsFound++
			skipped := false

			link, err := event.Locator("a").First().GetAttribute("href")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find link ... skipping\n")
				eventsErrors++
				continue
			}
			link = "https://www.gresham.ac.uk" + link
			title, err := event.Locator("a").First().InnerText()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find text ... skipping\n")
				eventsErrors++
				continue
			}
			fmt.Fprintf(os.Stderr, "Found '%s' at '%s'\n", title, link)

			// check for duplicate
			//
			for _, event := range allEvents {
				if event.Link == link {
					fmt.Fprintf(os.Stderr, "Duplicate event ... skipping\n")
					skipped = true
					continue
				}
			}
			if skipped {
				fmt.Fprintf(os.Stderr, "\n")
				continue
			}

			// categorize by description
			//

			// see if description is already cached, if so fetch
			// if not cached do a web query & classify
			//
			description := ""
			eventPrice := ""
			fetched := false
			var dt date.Date

			cacheEntry, err := eventCache.Get(link)
			if err != nil {
				start := time.Now()
				_, err = page2.Goto(link, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
				elapsed := time.Since(start)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not open '%s' ... skipping\n", link)
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}
				fmt.Fprintf(os.Stderr, "Done fetch page ... took %s\n", elapsed)

				paragraphs, err := page2.Locator(".sidebar__information--inner").Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).First().Locator("p").All()
				if err != nil || len(paragraphs) < 3 {
					fmt.Fprintf(os.Stderr, "Could not find date 1 ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				dateString, err := paragraphs[0].TextContent()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not find date 2 ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}
				timeString, err := paragraphs[1].TextContent()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not find time 3 ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}
				d := strings.NewReplacer("Date: ", "", "Time: ", "").Replace(dateString + " " + timeString) // yuk

				// parse date
				//
				dt, err = dateparser.Parse(defaultTime, d)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not parse date ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				description, err = page2.Locator(".m-entity__body").First().InnerText()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not read description '%s' ... skipping\n", link)
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					allEvents = append(allEvents, Event{Sort: dt.Time.Unix(), Name: title, Date: dt.Time.Local().Format("Mon 2 Jan 3:04PM"), Link: link, Categories: []string{"Link error"}, Include: false})
					continue
				}

				eventPrice = "£0.00"

				fetched = true
			} else {
				fmt.Fprintf(os.Stderr, "Used description from cache\n")

				description = cacheEntry.Description
				eventPrice = cacheEntry.Price
				title = cacheEntry.Title
				dt, _ = dateparser.Parse(defaultTime, cacheEntry.Date)
			}

			if !classify(title, description, link, eventPrice, dt.Time, cacheEntry, fetched || mustClassify || cliOptions.Reclassify) {
				greshamIncluded++
			}

		}

		ebPage++

		// click next page

		x, err := page1.GetByText(strconv.Itoa(ebPage), playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).All()
		if err != nil || len(x) < 1 {
			break
		}
		err = x[len(x)-1].Click(playwright.LocatorClickOptions{Timeout: playwright.Float(1000.0)})
		if err != nil {
			break
		}
	}
}

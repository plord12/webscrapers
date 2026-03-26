/**

find royal institution events

*/

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
	"github.com/playwright-community/playwright-go"
)

func rigb() {

	ebPage := 1

	// loop through all pages until we get nothing more ... store results in array for later sorting
	//
	url := "https://www.rigb.org/whats-on?type=6"

	fmt.Fprintf(os.Stderr, "Fetching %s\n", url)
	_, err := page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		_, err = page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not goto url: %v", err)
			return
		}
	}

	// reject cookie
	//
	page1.GetByText("SETTINGS", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})
	page1.GetByText("I Accept", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

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
			link = "https://www.rigb.org" + link
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

				d, err := page2.Locator(".datetime").First().GetAttribute("datetime")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				// parse date
				//
				dt, err = dateparser.Parse(nil, d)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not parse date ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				eventPrice, err = page2.Locator(".o-sidebar__price").First().InnerText()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not find eventPrice ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				description, err = page2.Locator(".m-entity__text").First().InnerText()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not read description '%s' ... skipping\n", link)
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					allEvents = append(allEvents, Event{Sort: dt.Time.Unix(), Name: title, Date: dt.Time.Local().Format("Mon 2 Jan 3:04PM"), Link: link, Categories: []string{"Link error"}, Include: false})
					continue
				}

				fetched = true
			} else {
				fmt.Fprintf(os.Stderr, "Used description from cache\n")

				description = cacheEntry.Description
				eventPrice = cacheEntry.Price
				title = cacheEntry.Title
				dt, _ = dateparser.Parse(nil, cacheEntry.Date)
			}

			if !classify(title, description, link, eventPrice, dt.Time, cacheEntry, fetched || mustClassify || cliOptions.Reclassify) {
				rigbIncluded++
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

/**

find university college london events

*/

package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
	"github.com/playwright-community/playwright-go"
)

func ucl() {

	ebPage := 1

	// loop through all pages until we get nothing more ... store results in array for later sorting
	//

	for {
		if ebPage > cliOptions.Maxpage {
			break
		}

		url := "https://www.ucl.ac.uk/events/all-events?ucl_audience[Public]=Public&ucl_event_type[Webinar]=Webinar&ucl_event_type[Online]=Online&page=" + strconv.Itoa(ebPage-1)

		fmt.Fprintf(os.Stderr, "Fetching %s\n", url)
		_, err := page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			_, err = page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not goto url: %v", err)
				fmt.Fprintf(os.Stderr, "\n")
				return
			}
		}

		// reject cookie
		//
		page1.GetByText("Accept necessary cookies", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

		events, err := page1.Locator(".event-feed-listing-item__content").Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).All()
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

				d, err := page2.Locator(".event-cta-card__paragraph--date-time").First().InnerText()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				// parse date
				//
				re := regexp.MustCompile(` to [0-9.pam]*`)
				d = re.ReplaceAllString(d, "")
				dt, err = dateparser.Parse(nil, d)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not parse date %s ... skipping\n", d)
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				eventPrice = "£0.00"

				description, err = page2.Locator(".basic-content__column").First().InnerText()

				fetched = true
			} else {
				fmt.Fprintf(os.Stderr, "Used description from cache\n")

				description = cacheEntry.Description
				eventPrice = cacheEntry.Price
				title = cacheEntry.Title
				dt, _ = dateparser.Parse(nil, cacheEntry.Date)
			}

			if !classify(title, description, link, eventPrice, dt.Time, cacheEntry, fetched || mustClassify || cliOptions.Reclassify) {
				uclIncluded++
			}
		}

		ebPage++
	}
}

/**

find university of york events

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

func york() {

	ebPage := 1

	// loop through all pages until we get nothing more ... store results in array for later sorting
	//
	url := "https://www.york.ac.uk/news-and-events/events/public-lectures/"

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
	page1.GetByText("Reject optional", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

	for {
		if ebPage > cliOptions.Maxpage {
			break
		}

		events, err := page1.Locator(".uoy_listing_item").Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).All()
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
			link = "https://www.york.ac.uk" + link
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

				keypoints, err := page2.Locator(".uoy_key_point_text").All()
				if err != nil || len(keypoints) < 3 {
					fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				d, err := keypoints[0].InnerText()
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
				dt, err = dateparser.Parse(defaultTime, d)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not parse date %s ... skipping\n", d)
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				eventPrice, err = keypoints[3].InnerText()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not find eventPrice ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}

				blocks, err := page2.Locator(".uoy_block_wrapper").All()
				if err != nil || len(blocks) < 4 {
					fmt.Fprintf(os.Stderr, "Could not find description ... skipping\n")
					fmt.Fprintf(os.Stderr, "\n")
					eventsErrors++
					continue
				}
				description, err = blocks[3].InnerText()
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
				dt, _ = dateparser.Parse(defaultTime, cacheEntry.Date)
			}

			if !classify(title, description, link, eventPrice, dt.Time, cacheEntry, fetched || mustClassify || cliOptions.Reclassify) {
				yorkIncluded++
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

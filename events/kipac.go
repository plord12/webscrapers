/**

find kipac events

*/

package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
	"github.com/playwright-community/playwright-go"
)

func kipac() {

	// This is a single page
	//
	url := "https://kipac.stanford.edu/events/upcoming-events"

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

	// load more events
	//
	page1.GetByText("Load More Events", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

	events, err := page1.Locator(".su-event-list-item__details").Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).All()
	if err != nil || len(events) == 0 {
		// no more pages
		return
	}

	for _, event := range events {
		eventsFound++
		skipped := false

		links, err := event.Locator("a").All()
		if err != nil || len(links) < 2 {
			fmt.Fprintf(os.Stderr, "Could not find link ... skipping\n")
			eventsErrors++
			continue
		}
		link, err := links[1].GetAttribute("href")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not find link ... skipping\n")
			eventsErrors++
			continue
		}
		link = "https://kipac.stanford.edu" + link
		title, err := links[1].InnerText()
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
			d, err := event.Locator("time").First().GetAttribute("datetime")

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

			eventPrice = "£0.00"

			description, err = event.Locator(".event-list_item__dek").First().InnerText(playwright.LocatorInnerTextOptions{Timeout: playwright.Float(2000.0)})

			fetched = true
		} else {
			fmt.Fprintf(os.Stderr, "Used description from cache\n")

			description = cacheEntry.Description
			eventPrice = cacheEntry.Price
			title = cacheEntry.Title
			dt, _ = dateparser.Parse(defaultTime, cacheEntry.Date)
		}

		if !classify(title, description, link, eventPrice, dt.Time, cacheEntry, fetched || mustClassify || cliOptions.Reclassify) {
			kipacIncluded++
		}

	}

}

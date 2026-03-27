/**

find university of Edinburgh events

*/

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/playwright-community/playwright-go"
)

func ed() {

	// loop through all pages until we get nothing more ... store results in array for later sorting
	//
	url := "https://tockify.com/edinburghuniversity/pinboard?tags=Lectures,%20Seminar,%20Inaugural-lectures"

	fmt.Fprintf(os.Stderr, "Fetching %s\n", url)
	_, err := page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})
	if err != nil {
		_, err = page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not goto url: %v", err)
			fmt.Fprintf(os.Stderr, "\n")
			return
		}
	}

	// load more events
	//
	page1.Locator(".btn-loadMore").Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})
	page1.Locator(".btn-loadMore").Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

	events, err := page1.Locator(".pincard").All()
	if err != nil || len(events) == 0 {
		fmt.Fprintf(os.Stderr, "No events %v %d\n", err, len(events))

		time.Sleep(30 * time.Second)

		// no more pages
		return
	}

	for _, event := range events {
		eventsFound++
		skipped := false

		link, err := event.Locator(".pincard__main__title").Locator("a").First().GetAttribute("href")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not find link ... skipping\n")
			eventsErrors++
			continue
		}
		link = "https://www.ed.ac.uk" + link
		cacheEntry, err := eventCache.Get(link)

		title, err := event.Locator(".pincard__main__title").Locator("a").First().InnerText()
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

		description, err := event.Locator(".pincard__main__preview").InnerText()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not find description ... skipping\n")
			eventsErrors++
			continue
		}

		date, err := event.Locator(".pincard__main__when").InnerText()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
			eventsErrors++
			continue
		}

		dt, err := dateparser.Parse(nil, date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse date %s ... skipping\n", date)
			fmt.Fprintf(os.Stderr, "\n")
			eventsErrors++
			continue
		}

		eventPrice := "£0.00"

		if !classify(title, description, link, eventPrice, dt.Time, cacheEntry, mustClassify || cliOptions.Reclassify) {
			edIncluded++
		}

	}
}

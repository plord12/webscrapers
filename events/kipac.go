/**

find kipac events

*/

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

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
		var categories []string
		description := ""
		eventPrice := ""
		fetched := false
		var dt date.Date

		cacheEntry, err := eventCache.Get(link)
		if err != nil {

			fmt.Fprintf(os.Stderr, "** datetime\n")
			d, err := event.Locator("time").First().GetAttribute("datetime")

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

			description, err = event.Locator(".event-list_item__dek").First().InnerText(playwright.LocatorInnerTextOptions{Timeout: playwright.Float(2000.0)})

			fetched = true
		} else {
			fmt.Fprintf(os.Stderr, "Used description from cache\n")

			description = cacheEntry.Description
			eventPrice = cacheEntry.Price
			title = cacheEntry.Title
			dt, _ = dateparser.Parse(nil, cacheEntry.Date)
		}

		if dt.Time.Before(startDate) || dt.Time.After(endDate) {
			fmt.Fprintf(os.Stderr, "Out of date range %s\n", dt.Time.Local().Format("Mon 2 Jan 3:04PM"))
			fmt.Fprintf(os.Stderr, "\n")
			continue
		}

		if fetched || mustClassify || cliOptions.Reclassify {

			fmt.Fprintf(os.Stderr, "Running classification\n")

			// classify by description
			//
			start := time.Now()
			limit := maxDescriptionWords
			words := strings.Split(title+" "+description, " ")
			if len(words) < limit {
				limit = len(words)
			}
			batch := []string{strings.Join(words[:limit], " ")}
			batchResult, err := classificationPipeline.RunPipeline(batch)
			if err != nil {
				panic(fmt.Sprintf("could not run pipeline: %v", err))
			}
			if len(batchResult.GetOutput()) == 1 {
				for i := range batchResult.ClassificationOutputs[0].SortedValues {
					if batchResult.ClassificationOutputs[0].SortedValues[i].Value > mlMinScore && i < maxCategoriesPerEvent {
						categories = append(categories, batchResult.ClassificationOutputs[0].SortedValues[i].Key)
					}
				}
			}

			// if no categories, perhaps try with whole description
			//
			if len(categories) == 0 {
				fmt.Fprintf(os.Stderr, "Running classification again\n")
				batch = []string{strings.Join(words, " ")}
				batchResult, err = classificationPipeline.RunPipeline(batch)
				if err != nil {
					panic(fmt.Sprintf("could not run pipeline: %v", err))
				}
				if len(batchResult.GetOutput()) == 1 {
					for i := range batchResult.ClassificationOutputs[0].SortedValues {
						if batchResult.ClassificationOutputs[0].SortedValues[i].Value > mlMinScore && i < maxCategoriesPerEvent {
							categories = append(categories, batchResult.ClassificationOutputs[0].SortedValues[i].Key)
						}
					}
				}
			}

			elapsed := time.Since(start)
			fmt.Fprintf(os.Stderr, "Done running pipeline ... took %s\n", elapsed)

			err = eventCache.Set(Cache{Title: title, Description: description, Date: dt.Time.Local().Format("Mon 2 Jan 3:04PM"), Categories: categories, Price: eventPrice}, link)
			if err != nil {
				panic(fmt.Sprintf("Could set cache %v", err))
			}

		} else {
			fmt.Fprintf(os.Stderr, "Used classification from cache\n")

			categories = cacheEntry.Categories
		}

		// add night time
		//
		if !cliOptions.Nighttime {
			if dt.Time.Hour() < nighttimeEndHour || dt.Time.Hour() > nighttimeStartHour {
				eventsSkippedByNightTime++
				categories = append(categories, "Night time")
				skipped = true
			}
		}

		for _, category := range categories {
			for _, exclude := range cliOptions.Exclude {
				if exclude == category {
					eventsSkippedByDescription++
					skipped = true
					break
				}
			}
			if skipped {
				break
			}
		}

		if skipped {
			allEvents = append(allEvents, Event{Sort: dt.Time.Unix(), Name: title, Date: dt.Time.Local().Format("Mon 2 Jan 3:04PM"), Link: link, Categories: categories, Include: false, Description: description, Price: eventPrice})
			fmt.Fprintf(os.Stderr, "Event excluded %s\n", strings.Join(categories, ","))
			fmt.Fprintf(os.Stderr, "\n")
			continue
		} else {
			kipacIncluded++
			allEvents = append(allEvents, Event{Sort: dt.Time.Unix(), Name: title, Date: dt.Time.Local().Format("Mon 2 Jan 3:04PM"), Link: link, Categories: categories, Include: true, Description: description, Price: eventPrice})
			fmt.Fprintf(os.Stderr, "Event included %s\n", strings.Join(categories, ","))
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

}

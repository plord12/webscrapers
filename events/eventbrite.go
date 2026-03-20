/**

find eventbrite events

*/

package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
	"github.com/playwright-community/playwright-go"
)

func eventbrite() {

	for _, price := range []string{"free", "paid"} {

		ebPage := 1

		// loop through all pages until we get nothing more ... store results in array for later sorting
		//
		for {
			if ebPage > cliOptions.Maxpage {
				break
			}

			var url string
			// https://www.eventbrite.com/d/online/free--science-and-tech--events/?page=1&start_date=2026-03-09&end_date=2026-03-23&lang=en

			if cliOptions.StartDate != "" {
				url = "https://www.eventbrite.com/d/online/" + price + "--" + cliOptions.Category + "--events/?page=" + strconv.Itoa(ebPage) + "&start_date=" + startDate.Format("2006-01-02") + "&end_date=" + endDate.Format("2006-01-02") + "&lang=en"
			} else {
				url = "https://www.eventbrite.com/d/online/" + price + "--" + cliOptions.Category + "--events--" + cliOptions.Date + "/?page=" + strconv.Itoa(ebPage) + "&lang=en"
			}
			fmt.Fprintf(os.Stderr, "Fetching %s\n", url)
			_, err := page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			if err != nil {
				_, err = page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
				if err != nil {
					panic(fmt.Sprintf("could not goto url: %v", err))
				}
			}

			// reject cookie
			//
			//page.GetByText("Reject all", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

			events, err := page1.Locator(".event-card-details").Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).All()
			if err != nil {
				panic("Could not find links")
			}

			if len(events) == 0 {
				// no more pages
				break
			}

			// do our best to wait for rendering
			//
			for i := 0; i < 10; i++ {
				firstEventParagraphs, err := page1.Locator(".event-card-details").First().Locator("p").All()
				if err == nil {
					found := false
					if len(firstEventParagraphs) > 1 {
						for j := 0; j < len(firstEventParagraphs)-1; j++ {
							t, _ := firstEventParagraphs[j].TextContent()
							if strings.Contains(t, "•") {
								found = true
							}
						}
						if found {
							break
						}
					}
				}
				time.Sleep(1 * time.Second)
			}

			for _, event := range events {
				eventsFound++
				skipped := false

				link, err := event.Locator(".event-card-link").First().GetAttribute("href")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not find link ... skipping\n")
					eventsErrors++
					continue
				}
				title, err := event.Locator(".event-card-link").First().TextContent()
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

				paragraphs, err := event.Locator("p").All()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
					eventsErrors++
					continue
				}

				// need to look at all paragraphs looking for a date & price
				//
				var dt date.Date
				found := false
				for _, para := range paragraphs {
					t, _ := para.TextContent()

					// parse date into sort key
					//
					d := strings.NewReplacer("  ", " ",
						" • ", " ",
						", ", " ").Replace(t)
					re := regexp.MustCompile(`\+ [0-9]* more`)
					d = re.ReplaceAllString(d, "")
					dt, err = dateparser.Parse(nil, d)
					if err != nil {
						continue
					}
					if dt.Time.Before(time.Now()) {
						dt.Time = dt.Time.AddDate(0, 0, 7)
					}
					found = true
					break
				}

				if !found {
					fmt.Fprintf(os.Stderr, "Could not parse date ... skipping\n")
					eventsErrors++
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
				localPrice := 0.0
				fetched := false
				cacheEntry, err := eventCache.Get(link)
				if err != nil {
					start := time.Now()
					_, err = page2.Goto(link, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
					elapsed := time.Since(start)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Could not open '%s' ... skipping\n", link)
						eventsErrors++
						continue
					}
					fmt.Fprintf(os.Stderr, "Done fetch page ... took %s\n", elapsed)

					err = page2.GetByTestId("section-wrapper-overview").WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(1000)})
					if err == nil {
						description, _ = page2.GetByTestId("section-wrapper-overview").First().InnerText()
					} else {
						err = page2.GetByTestId("overview").WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(1000)})
						if err == nil {
							description, _ = page2.GetByTestId("overview").First().InnerText()
						} else {
							fmt.Fprintf(os.Stderr, "Could not read '%s' ... skipping\n", link)
							eventsErrors++
							allEvents = append(allEvents, Event{Sort: dt.Time.Unix(), Name: title, Date: dt.Time.Local().Format("Mon 2 Jan 3:04PM"), Link: link, Categories: []string{"Link error"}, Include: false})
							continue
						}
					}

					// convert to local price
					//
					err = page2.GetByTestId("conversion-bar-headline").WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(1000)})
					if err == nil {
						eventPrice, _ = page2.GetByTestId("conversion-bar-headline").First().InnerText()
						localPrice, err = convertToGBP(eventPrice)
						if err == nil {
							eventPrice = fmt.Sprintf("£%.2f", localPrice)
						}
					}

					/*
						// could be more reliable than main page
						eventDate, err := page2.GetByTestId("conversion-bar-date").First().InnerText()
						if err == nil {
							fmt.Fprintf(os.Stderr, "Date: %s\n", eventDate)
						}
					*/

					fetched = true
				} else {
					fmt.Fprintf(os.Stderr, "Used description from cache\n")

					description = cacheEntry.Description
					eventPrice = cacheEntry.Price
					title = cacheEntry.Title
					localPrice, err = convertToGBP(eventPrice)
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

					err = eventCache.Set(Cache{Title: title, Description: description, Categories: categories, Price: eventPrice}, link)
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

				// add expensive
				//
				if localPrice >= cliOptions.MaxPrice {
					eventsSkippedByPrice++
					categories = append(categories, "Expensive")
					skipped = true
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
					fmt.Fprintf(os.Stderr, "Event excluded %s\n\n", strings.Join(categories, ","))
					continue
				} else {
					allEvents = append(allEvents, Event{Sort: dt.Time.Unix(), Name: title, Date: dt.Time.Local().Format("Mon 2 Jan 3:04PM"), Link: link, Categories: categories, Include: true, Description: description, Price: eventPrice})
					fmt.Fprintf(os.Stderr, "Event included %s\n\n", strings.Join(categories, ","))
				}
			}

			ebPage++
		}
	}
}

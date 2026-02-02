/**

find eventbrite events, output in wordpress format

*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless  bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Category  string `short:"c" long:"category" description:"Category" default:"science-and-tech" env:"CATEGORY"`
	Date      string `short:"d" long:"date" description:"Date" default:"next-month" env:"DATE"`
	Price     string `short:"p" long:"price" description:"Price" default:"free" env:"PRICE"`
	Nighttime bool   `short:"n" long:"nighttime" description:"Include nighttime events" env:"NIGHTTIME"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)

func main() {

	// parse flags
	//
	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}

	// setup
	//
	page := utils.StartChromium(options.Headless)
	defer utils.Finish(page)

	// main page
	//
	// FIX THIS - allow multiple passes & remove duplicates
	//

	type Event struct {
		Sort int64
		Name string
		Date string
		Link string
	}
	var listEvents []Event

	ebPage := 1

	// loop through all pages until we get nothing more ... store results in array for later sorting
	//
	for {
		url := "https://www.eventbrite.com/d/online/" + options.Price + "--" + options.Category + "--events--" + options.Date + "/?page=" + strconv.Itoa(ebPage) + "&lang=en"
		fmt.Fprintf(os.Stderr, "Fetching %s\n", url)
		_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			panic(fmt.Sprintf("could not goto url: %v", err))
		}

		events, err := page.Locator(".event-card-details").Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).All()
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
			firstEventParagraphs, err := page.Locator(".event-card-details").First().Locator("p").All()
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
			link, err := event.Locator(".event-card-link").First().GetAttribute("href")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find link ... skipping\n")
				break
			}
			text, err := event.Locator(".event-card-link").First().TextContent()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find text ... skipping\n")
				break
			}
			// fmt.Fprintf(os.Stderr, "Found '%s' at '%s'\n", text, link)

			paragraphs, err := event.Locator("p").All()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
				break
			}

			date := ""
			if len(paragraphs) == 2 {
				date, err = paragraphs[0].TextContent()
			}
			if len(paragraphs) == 3 {
				date, err = paragraphs[1].TextContent()
			}
			if date == "" {
				fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
				break
			}

			// parse date into sort key
			//
			d := strings.ReplaceAll(strings.ReplaceAll(date, "  ", " "), " • ", " ")
			d = strings.Join(strings.Split(d, " ")[0:5], " ")
			t, err := time.Parse("Mon, Jan 2 3:04 PM", d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not parse date '%s' ... skipping\n", d)
				break
			}
			t = t.AddDate(time.Now().Year(), 0, 0)
			listEvents = append(listEvents, Event{Sort: t.Unix(), Name: text, Date: d, Link: link})
		}

		ebPage++
	}

	// sort, filter & display
	//
	fmt.Printf("eventbrite has been run with the following options :\n")
	fmt.Printf("	Headless=%v\n", options.Headless)
	fmt.Printf("	Category=%s\n", options.Category)
	fmt.Printf("	Date=%s\n", options.Date)
	fmt.Printf("	Price=%s\n", options.Price)
	fmt.Printf("	Nighttime=%v\n", options.Nighttime)
	fmt.Printf("\n")
	fmt.Printf("Below is generated wordpress source which can be cut&pasted onto your page.\n")
	fmt.Printf("\n")

	fmt.Printf("<!-- wp:list --><ul class=\"wp-block-list\">\n")
	sort.Slice(listEvents, func(i, j int) bool {
		return listEvents[i].Sort < listEvents[j].Sort
	})
	for _, event := range listEvents {
		if !options.Nighttime {
			if time.Unix(event.Sort, 0).Hour() < 8 || time.Unix(event.Sort, 0).Hour() > 20 {
				break
			}
		}
		fmt.Printf("<!-- wp:list-item -->\n")
		fmt.Printf("<li>%s <a href=\"%s\">%s</a></li>\n", event.Date, event.Link, event.Name)
		fmt.Printf("<!-- /wp:list-item -->\n")
	}
	fmt.Printf("</ul><!-- /wp:list -->\n")

	bufio.NewWriter(os.Stdout).Flush()
}

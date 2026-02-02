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
	Maxpage   int    `short:"m" long:"maxpage" description:"Max page number to fetch" default:"1000" env:"MAXPAGE"`
	Format    string `short:"f" long:"format" description:"Format - list or table" default:"list" env:"FORMAT"`
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
		if ebPage > options.Maxpage {
			break
		}

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
				continue
			}
			text, err := event.Locator(".event-card-link").First().TextContent()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find text ... skipping\n")
				continue
			}
			// fmt.Fprintf(os.Stderr, "Found '%s' at '%s'\n", text, link)

			paragraphs, err := event.Locator("p").All()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
				continue
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
				continue
			}

			// parse date into sort key
			//
			d := strings.ReplaceAll(strings.ReplaceAll(date, "  ", " "), " • ", " ")
			d = strings.Join(strings.Split(d, " ")[0:5], " ")
			t, err := time.Parse("Mon, Jan 2 3:04 PM", d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not parse date '%s' ... skipping\n", d)
				continue
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
	fmt.Printf("	Maxpage=%d\n", options.Maxpage)
	fmt.Printf("	Format=%s\n", options.Format)
	fmt.Printf("\n")
	fmt.Printf("Below is generated wordpress source which can be cut&pasted onto your page.\n")
	fmt.Printf("Switch to the `Code editor` (top right menu), paste then switch back to `Visual editor`.\n")
	fmt.Printf("\n")

	if options.Format == "list" {
		fmt.Printf("<!-- wp:list --><ul class=\"wp-block-list\">\n")
	} else {
		fmt.Printf("<!-- wp:table {\"hasFixedLayout\":false,\"align\":\"left\",\"className\":\"is-style-regular\"} -->\n")
		fmt.Printf("<figure class=\"wp-block-table alignleft is-style-regular\">\n")
		fmt.Printf("<table><thead><tr><th>Date</th><th>Event &amp; Link</th></tr></thead><tbody>\n")
	}
	sort.Slice(listEvents, func(i, j int) bool {
		return listEvents[i].Sort < listEvents[j].Sort
	})
	for _, event := range listEvents {
		if !options.Nighttime {
			if time.Unix(event.Sort, 0).Hour() < 8 || time.Unix(event.Sort, 0).Hour() > 20 {
				continue
			}
		}
		if options.Format == "list" {
			fmt.Printf("<!-- wp:list-item -->\n")
			fmt.Printf("<li>%s <a href=\"%s\">%s</a></li>\n", event.Date, event.Link, event.Name)
			fmt.Printf("<!-- /wp:list-item -->\n")
		} else {
			fmt.Printf("<tr><td>%s</td><td><a href=\"%s\">%s</a></td></tr>\n", event.Date, event.Link, event.Name)
		}
	}
	if options.Format == "list" {
		fmt.Printf("</ul><!-- /wp:list -->\n")
	} else {
		fmt.Printf("</tbody></table></figure>\n")
		fmt.Printf("<!-- /wp:table -->\n")
	}

	bufio.NewWriter(os.Stdout).Flush()
}

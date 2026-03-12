/**

find eventbrite events, output in wordpress format

*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gildas/go-cache"
	"github.com/jessevdk/go-flags"
	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/backends"
	"github.com/knights-analytics/hugot/pipelines"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/markusmobius/go-dateparser"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless  bool     `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Category  string   `short:"c" long:"category" description:"Category" default:"science-and-tech" env:"CATEGORY"`
	Date      string   `short:"d" long:"date" description:"Date" default:"next-month" default:"" env:"DATE"`
	StartDate string   `short:"s" long:"startdate" description:"Start date (YYYY-MM-DD)" env:"STARTDATE"`
	EndDate   string   `short:"a" long:"enddate" description:"End date (YYYY-MM-DD)" env:"ENDDATE"`
	MaxPrice  float64  `short:"p" long:"maxprice" description:"Max price for event (£)" default:"20" env:"PRICE"`
	Nighttime bool     `short:"n" long:"nighttime" description:"Include nighttime events" env:"NIGHTTIME"`
	Maxpage   int      `short:"m" long:"maxpage" description:"Max page number to fetch" default:"1000" env:"MAXPAGE"`
	Format    string   `short:"f" long:"format" description:"Format - list, table or tablepress" default:"list" choice:"list" choice:"table" choice:"tablepress" env:"FORMAT"`
	Include   []string `short:"i" long:"include" description:"Include - list of categories to include" env:"INCLUDE"`
	Exclude   []string `short:"x" long:"exclude" description:"Exclude - list of categories to exclude" env:"EXCLUDE"`
	Clear     bool     `short:"z" long:"clear" description:"Clear the cache ... eg change in categories" env:"CLEAR"`
	Save      string   `short:"v" long:"save" description:"Filename to save output to" env:"SAVE"`
}

var cliOptions Options
var parser = flags.NewParser(&cliOptions, flags.Default)

type Event struct {
	Sort        int64
	Name        string
	Date        string
	Price       string
	Link        string
	Categories  []string
	Description string
	Include     bool
}

var allEvents []Event

// currency conversions
var currencyUKP, _ = regexp.Compile(".?£([0-9.]*)[^0-9]*")
var currencyUSD, _ = regexp.Compile(".?\\$([0-9.]*)[^0-9]*")
var currencyAUD, _ = regexp.Compile(".?A\\$([0-9.]*)[^0-9]*")
var currencyEUR, _ = regexp.Compile(".?€([0-9.]*)[^0-9]*")
var currencyCAD, _ = regexp.Compile(".?CA\\$([0-9.]*)[^0-9]*")
var currencySGD, _ = regexp.Compile(".?SGD.?([0-9.]*)[^0-9]*")
var currencyARS, _ = regexp.Compile(".?ARS.?([0-9.]*)[^0-9]*")

type Rate struct {
	Code string
	Rate float64
}

var rates map[string]Rate

// constants
//
// # Machine Learning
//
// GoMLX: unimplemented ONNX op "ReduceSum" in Node "/model/ReduceSum" [ReduceSum](/model/Cast_output_0, onnx::ReduceSum_2901) -> /model/ReduceSum_output_0 - attrs[keepdims (INT)]
// ORT: index out of range [-1]
// var mlModelFile = "onnx/model.onnx"
// var mlModelFile = "knowledgator/gliclass-small-v1.0"
//
// GoMLX: DotGeneral contracting dimensions don't match: lhs[2]=1 != rhs[0]=576
// ORT: index out of range [-1]
// var mlModelFile = "model.onnx"
// var mlModelFile = "cnmoro/gliclass-edge-v3.0-onnx"
//
// var mlModelFile = "KnightsAnalytics/distilbert-base-uncased-finetuned-sst-2-english"
//
// ORT: invalid memory address or nil pointer dereference
// var mlModelFile = winado/gliclass-base-onnx"
//
// works with ORT & XLA
// 2.936602584s ORT
var mlModelFile = "onnx/model.onnx"
var mlModel = "MoritzLaurer/deberta-v3-large-zeroshot-v2.0"

// 957.691292ms ORT
// var mlModelFile = "MoritzLaurer/deberta-v3-base-zeroshot-v2.0"
//
// fails
// var mlModelFile = "onnx-community/deberta-v3-small"
//
// 957.295666ms ORT, hang GO, 4.726374958s XLA
// const mlModel = "KnightsAnalytics/deberta-v3-base-zeroshot-v1"
// const mlModelFile = "model.onnx"
const mlBackend = "ORT"
const mlMinScore = 0.1

// night time
const nighttimeEndHour = 8
const nighttimeStartHour = 22

const maxCategoriesPerEvent = 3
const maxDescription = 250

var palette []colorful.Color

func main() {

	// parse flags
	//
	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}

	getExchangeRates()

	// validate arguments
	//
	var startDate time.Time
	var endDate time.Time

	if cliOptions.StartDate != "" {
		dt, err := dateparser.Parse(nil, cliOptions.StartDate)
		if err != nil {
			panic(fmt.Sprintf("could not parse start date %s: %v", cliOptions.StartDate, err))
		}
		startDate = dt.Time

		dt, err = dateparser.Parse(nil, cliOptions.EndDate)
		if err != nil {
			panic(fmt.Sprintf("could not parse end date %s: %v", cliOptions.EndDate, err))
		}
		endDate = dt.Time
	}

	// if found in the cach, must still re-classify since categories have changed
	//
	mustClassify := false

	// disk cache ... perhaps this should be the same as Event ?
	//

	type Cache struct {
		Description string
		Price       string
		Categories  []string
	}
	cache := cache.New[Cache]("eventbrite", cache.CacheOptionPersistent).WithExpiration(7 * 24 * time.Hour)
	if cliOptions.Clear {
		cache.Clear()
	} else {
		lastRuncategories, err := cache.Get("all categories")
		if err == nil && reflect.DeepEqual(lastRuncategories.Categories, append(cliOptions.Include, cliOptions.Exclude...)) {
			// all good
		} else {
			mustClassify = true
			fmt.Fprintf(os.Stderr, "Categories have changed, have to re-run classifications\n")
		}
	}
	cache.Set(Cache{Categories: append(cliOptions.Include, cliOptions.Exclude...)}, "all categories")

	// FIX THIS - add https://www.gresham.ac.uk/watch-now Gresham
	// FIX THIS - add https://www.rigb.org/whats-on?see-all Royal institution but only the online and there are fees. Booking though Eventbrite
	// FIX THIS - add https://www.york.ac.uk/news-and-events/events/  Uni of York online variable
	// FIX THIS - add https://www.ucl.ac.uk/events/all-events UCL online variable
	// FIX THIS - add https://www.linnean.org/meetings-and-events Linnean Society two or three
	// FIX THIS - add https://www.bcs.org/events-calendar/ BCS (the Chartered Institute for IT) several hybrid or webinar items each month. Booked through Eventbrite. But not all appear under science and tech
	// FIX THIS - add https://kipac.stanford.edu/events/upcoming-events KIPAC (Kavli Institute for particle Astrophysics and cosmology) Stanford University several items each month
	// FIX THIS - see if filtering based on English first would work out better
	// FIX THIS - consider adding main search page to cache
	// FIX THIS - re-try if page not found

	// stats
	//
	eventsFound := 0
	eventsSkippedByDescription := 0
	eventsSkippedByNightTime := 0
	eventsSkippedByPrice := 0
	eventsErrors := 0

	// machine learning classification
	//
	var session *hugot.Session
	if mlBackend == "XLA" {
		session, err = hugot.NewXLASession()
	} else if mlBackend == "ORT" {
		//session, err = hugot.NewORTSession(options.WithCoreML(map[string]string{"ModelFormat": "MLProgram", "MLComputeUnits": "ALL", "RequireStaticInputShapes": "0", "EnableOnSubgraphs": "0"}))
		session, err = hugot.NewORTSession()

	} else {
		// tends to hang
		session, err = hugot.NewGoSession()
	}
	if err != nil {
		panic(fmt.Sprintf("Could not start hugot: %v", err))
	}
	defer func(session *hugot.Session) {
		err := session.Destroy()
		if err != nil {
			panic("Could not destroy hugot")
		}
	}(session)
	downloadOptions := hugot.NewDownloadOptions()
	downloadOptions.OnnxFilePath = mlModelFile
	modelPath, err := hugot.DownloadModel(mlModel, "./models/", downloadOptions)
	if err != nil {
		panic(fmt.Sprintf("could not download model: %v", err))
	}
	config := hugot.ZeroShotClassificationConfig{
		ModelPath: modelPath,
		Name:      "testPipeline",
		Options: []backends.PipelineOption[*pipelines.ZeroShotClassificationPipeline]{
			pipelines.WithLabels(append(cliOptions.Include, cliOptions.Exclude...)),
			pipelines.WithMultilabel(false),
		},
	}
	classificationPipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		panic(fmt.Sprintf("could not create pipeline: %v", err))
	}

	// test classification
	//
	/*
		fmt.Fprintf(os.Stderr, "Running pipeline\n")
		batch := []string{"Thermal Vision for Bats: Practical Applications in Ecology"}
		start := time.Now()
		batchResult, err := classificationPipeline.RunPipeline(batch)
		elapsed := time.Since(start)
		if err != nil {
			panic(fmt.Sprintf("could not run pipeline: %v", err))
		}
		fmt.Fprintf(os.Stderr, "Done running pipeline ... took %s\n", elapsed)
		if len(batchResult.GetOutput()) == 1 {
			for i := range batchResult.ClassificationOutputs[0].SortedValues {
				if batchResult.ClassificationOutputs[0].SortedValues[i].Value > 0.2 {
					fmt.Fprintf(os.Stderr, "%s ", batchResult.ClassificationOutputs[0].SortedValues[i].Key)
				}
			}
			fmt.Fprintf(os.Stderr, "\n")
		}
		panic("done")
	*/

	// setup
	//
	page := utils.StartCamoufox(cliOptions.Headless)
	defer utils.Finish(page)

	newContext, err := page.Context().Browser().NewContext()
	if err != nil {
		panic(fmt.Sprintf("could not open new page: %v", err))
	}
	page2, err := newContext.NewPage()
	if err != nil {
		panic(fmt.Sprintf("could not open new page: %v", err))
	}
	defer utils.Finish(page2)

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
			_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			if err != nil {
				_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
				if err != nil {
					panic(fmt.Sprintf("could not goto url: %v", err))
				}
			}

			// reject cookie
			//
			//page.GetByText("Reject all", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

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

				date := ""
				if len(paragraphs) == 2 {
					date, err = paragraphs[0].TextContent()
				}
				if len(paragraphs) == 3 {
					date, err = paragraphs[1].TextContent()
				}
				if date == "" {
					fmt.Fprintf(os.Stderr, "Could not find date ... skipping\n")
					eventsErrors++
					continue
				}

				// parse date into sort key
				//
				d := strings.NewReplacer("  ", " ",
					" • ", " ",
					", ", " ").Replace(date)
				re := regexp.MustCompile(`\+ [0-9]* more`)
				d = re.ReplaceAllString(d, "")
				dt, err := dateparser.Parse(nil, d)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not parse date '%s' %v ... skipping\n", d, err)
					eventsErrors++
					continue
				}
				if dt.Time.Before(time.Now()) {
					dt.Time = dt.Time.AddDate(0, 0, 7)
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
				cacheEntry, err := cache.Get(link)
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
					localPrice, err = convertToGBP(eventPrice)
				}

				//fmt.Fprintf(os.Stderr, "Description '%s'\n", description)

				if fetched || mustClassify {

					fmt.Fprintf(os.Stderr, "Running classification\n")

					// classify by description
					//
					limit := maxDescription
					totalText := title + " " + description
					if len(totalText) < limit {
						limit = len(totalText)
					}
					batch := []string{totalText[:limit]}

					start := time.Now()
					batchResult, err := classificationPipeline.RunPipeline(batch)
					elapsed := time.Since(start)
					if err != nil {
						panic(fmt.Sprintf("could not run pipeline: %v", err))
					}
					fmt.Fprintf(os.Stderr, "Done running pipeline ... took %s\n", elapsed)

					if len(batchResult.GetOutput()) == 1 {
						for i := range batchResult.ClassificationOutputs[0].SortedValues {
							if batchResult.ClassificationOutputs[0].SortedValues[i].Value > mlMinScore && i < maxCategoriesPerEvent {
								categories = append(categories, batchResult.ClassificationOutputs[0].SortedValues[i].Key)
							}
						}
					}

					err = cache.Set(Cache{Description: description, Categories: categories, Price: eventPrice}, link)
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

	fmt.Printf("eventbrite has been run with the following options :\n")
	fmt.Printf("	Headless=%v\n", cliOptions.Headless)
	fmt.Printf("	Category=%s\n", cliOptions.Category)
	fmt.Printf("	Date=%s\n", cliOptions.Date)
	fmt.Printf("	Nighttime=%v\n", cliOptions.Nighttime)
	fmt.Printf("	Maxpage=%d\n", cliOptions.Maxpage)
	fmt.Printf("	Format=%s\n", cliOptions.Format)
	fmt.Printf("	Max price=£%.2f\n", cliOptions.MaxPrice)
	fmt.Printf("	Include=%s\n", strings.Join(cliOptions.Include, ","))
	fmt.Printf("	Exclude=%s\n", strings.Join(cliOptions.Exclude, ","))
	fmt.Printf("	Machine learning model %s with %s backend\n", mlModel, mlBackend)
	fmt.Printf("\n")
	fmt.Printf("There were %d events found.  Of which :\n", eventsFound)
	fmt.Printf("	%d were skipped due to excluded categories match\n", eventsSkippedByDescription)
	fmt.Printf("	%d were skipped due to nighttime\n", eventsSkippedByNightTime)
	fmt.Printf("	%d were skipped due to high price\n", eventsSkippedByPrice)
	fmt.Printf("	%d errors\n", eventsErrors)
	fmt.Printf("\n")

	// sort
	//
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Sort < allEvents[j].Sort
	})

	// and generate
	//

	palette, err = colorful.HappyPalette(len(cliOptions.Include) + len(cliOptions.Exclude))
	if err != nil {
		panic(fmt.Sprintf("could not generate colors: %v", err))
	}

	fmt.Printf("Colour palette is:\n")

	for i, category := range append(cliOptions.Include, cliOptions.Exclude...) {
		fmt.Printf("	%s - <mark style=\"background-color:%s\" class=\"has-inline-color has-white-color\"> %s </mark>\n", category, palette[i].Hex(), category)
	}
	fmt.Printf("\n")

	report := ""
	if cliOptions.Format == "list" {

		fmt.Printf("Below is generated wordpress source which can be cut&pasted onto your page.\n")
		fmt.Printf("Switch to the `Code editor` (top right menu), paste then switch back to `Visual editor`.\n")
		fmt.Printf("\n")

		report = generateList()

	} else if cliOptions.Format == "table" {

		fmt.Printf("Below is generated wordpress source which can be cut&pasted onto your page.\n")
		fmt.Printf("Switch to the `Code editor` (top right menu), paste then switch back to `Visual editor`.\n")
		fmt.Printf("\n")

		report = generateTable()

	} else {

		fmt.Printf("Below is generated tablepress in json.  Use the TablePress menu in wordpress to import\n")
		fmt.Printf("this to a new table and then add that TablePress to your page.\n")
		fmt.Printf("\n")

		report = generateTablePress()
	}

	if len(cliOptions.Save) == 0 {
		fmt.Print(report)
	} else {
		fi, err := os.Create(cliOptions.Save)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := fi.Close(); err != nil {
				panic(err)
			}
		}()
		fi.WriteString(report)
	}
}

func getExchangeRates() {
	// should cache this
	httpClient := http.Client{Timeout: time.Second * 2}

	req, err := http.NewRequest(http.MethodGet, "https://www.floatrates.com/daily/gbp.json", nil)
	if err != nil {
		panic(fmt.Sprintf("could not create http session: %v", err))

	}
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		panic(fmt.Sprintf("could not get exchange rates from floatrates: %v", err))

	}
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		panic(fmt.Sprintf("could not read exchange rates from floatrates: %v", err))
	}
	jsonErr := json.Unmarshal(body, &rates)
	if jsonErr != nil {
		panic(fmt.Sprintf("could not process exchange rates json: %v", err))
	}
}

func convertToGBP(currencyString string) (float64, error) {

	if currencyString == "Free" {
		return 0, nil
	}
	ukp := currencyUKP.FindStringSubmatch(currencyString)
	if len(ukp) > 0 {
		converted, err := strconv.ParseFloat(ukp[1], 32)
		return converted, err
	}
	cad := currencyCAD.FindStringSubmatch(currencyString)
	if len(cad) > 0 {
		converted, err := strconv.ParseFloat(cad[1], 32)
		if err == nil {
			return converted / rates["cad"].Rate, nil
		}
		return converted, err
	}
	usd := currencyUSD.FindStringSubmatch(currencyString)
	if len(usd) > 0 {
		converted, err := strconv.ParseFloat(usd[1], 32)
		if err == nil {
			return converted / rates["usd"].Rate, nil
		}
		return converted, err
	}
	aud := currencyAUD.FindStringSubmatch(currencyString)
	if len(aud) > 0 {
		converted, err := strconv.ParseFloat(aud[1], 32)
		if err == nil {
			return converted / rates["aud"].Rate, nil
		}
		return converted, err
	}
	eur := currencyEUR.FindStringSubmatch(currencyString)
	if len(eur) > 0 {
		converted, err := strconv.ParseFloat(eur[1], 32)
		if err == nil {
			return converted / rates["eur"].Rate, nil
		}
		return converted, err
	}
	sgd := currencySGD.FindStringSubmatch(currencyString)
	if len(sgd) > 0 {
		converted, err := strconv.ParseFloat(sgd[1], 32)
		if err == nil {
			return converted / rates["sgd"].Rate, nil
		}
		return converted, err
	}
	ars := currencyARS.FindStringSubmatch(currencyString)
	if len(ars) > 0 {
		converted, err := strconv.ParseFloat(ars[1], 32)
		if err == nil {
			return converted / rates["ars"].Rate, nil
		}
		return converted, err
	}
	return 0, errors.New("failed to parse")
}

func generateList() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "<!-- wp:heading -->\n")
	if cliOptions.StartDate != "" {
		fmt.Fprintf(&sb, "<h2 class=\"wp-block-heading\">%s to %s</h2>\n", cliOptions.StartDate, cliOptions.EndDate)
	} else {
		fmt.Fprintf(&sb, "<h2 class=\"wp-block-heading\">%s</h2>\n", cliOptions.Date)
	}
	fmt.Fprintf(&sb, "<!-- /wp:heading -->\n")

	fmt.Fprintf(&sb, "<!-- wp:list --><ul class=\"wp-block-list\">\n")

	for _, event := range allEvents {
		if event.Include {
			fmt.Fprintf(&sb, "<!-- wp:list-item -->\n")
			fmt.Fprintf(&sb, "<li>%s (%s) ", event.Date, event.Price)
			for _, category := range event.Categories {
				// find a color
				color := palette[0]
				for i, cat := range append(cliOptions.Include, cliOptions.Exclude...) {
					if cat == category {
						color = palette[i%len(palette)]
						break
					}
				}
				fmt.Fprintf(&sb, "<mark style=\"background-color:%s\" class=\"has-inline-color has-white-color\"> %s </mark> ", color.Hex(), category)
			}
			fmt.Fprintf(&sb, "<a href=\"%s\">%s</a></li>\n", event.Link, html.EscapeString(event.Name))
			fmt.Fprintf(&sb, "<!-- /wp:list-item -->\n")
		}
	}

	fmt.Fprintf(&sb, "<!-- wp:heading -->\n")
	fmt.Fprintf(&sb, "<h2 class=\"wp-block-heading\">Excluded</h2>\n")
	fmt.Fprintf(&sb, "<!-- /wp:heading -->\n")

	fmt.Fprintf(&sb, "</ul><!-- /wp:list -->\n")
	fmt.Fprintf(&sb, "<!-- wp:list --><ul class=\"wp-block-list\">\n")

	for _, event := range allEvents {
		if !event.Include {
			fmt.Fprintf(&sb, "<!-- wp:list-item -->\n")
			fmt.Fprintf(&sb, "<li>%s (%s) ", event.Date, event.Price)
			for _, category := range event.Categories {
				// find a color
				color := palette[0]
				for i, cat := range append(cliOptions.Include, cliOptions.Exclude...) {
					if cat == category {
						color = palette[i%len(palette)]
						break
					}
				}
				fmt.Fprintf(&sb, "<mark style=\"background-color:%s\" class=\"has-inline-color has-white-color\"> %s </mark> ", color.Hex(), category)
			}
			fmt.Fprintf(&sb, "<a href=\"%s\">%s</a></li>\n", event.Link, html.EscapeString(event.Name))
			fmt.Fprintf(&sb, "<!-- /wp:list-item -->\n")
		}
	}

	fmt.Fprintf(&sb, "</ul><!-- /wp:list -->\n")

	return sb.String()
}

func generateTable() string {
	var sb strings.Builder

	// sort & display
	//

	fmt.Printf("<!-- wp:heading -->\n")
	if cliOptions.StartDate != "" {
		fmt.Fprintf(&sb, "<h2 class=\"wp-block-heading\">%s to %s</h2>\n", cliOptions.StartDate, cliOptions.EndDate)
	} else {
		fmt.Fprintf(&sb, "<h2 class=\"wp-block-heading\">%s</h2>\n", cliOptions.Date)
	}
	fmt.Fprintf(&sb, "<!-- /wp:heading -->\n")

	fmt.Fprintf(&sb, "<!-- wp:table {\"hasFixedLayout\":false,\"align\":\"left\",\"className\":\"is-style-regular\"} -->\n")
	fmt.Fprintf(&sb, "<figure class=\"wp-block-table alignleft is-style-regular\">\n")
	fmt.Fprintf(&sb, "<table><thead><tr><th>Date</th><th>Price</th><th>Categories</th><th>Event &amp; Link</th></tr></thead><tbody>\n")

	for _, event := range allEvents {
		if event.Include {
			fmt.Fprintf(&sb, "<tr><td>%s</td><td>%s</td><td>", event.Date, event.Price)
			for _, category := range event.Categories {
				// find a color
				color := palette[0]
				for i, cat := range append(cliOptions.Include, cliOptions.Exclude...) {
					if cat == category {
						color = palette[i%len(palette)]
						break
					}
				}
				fmt.Fprintf(&sb, "<mark style=\"background-color:%s\" class=\"has-inline-color has-white-color\"> %s </mark> ", color.Hex(), category)
			}
			fmt.Fprintf(&sb, "</td><td><a href=\"%s\">%s</a></td></tr>\n", event.Link, html.EscapeString(event.Name))
		}
	}

	fmt.Fprintf(&sb, "</tbody></table></figure>\n")
	fmt.Fprintf(&sb, "<!-- /wp:table -->\n")

	fmt.Fprintf(&sb, "<!-- wp:heading -->\n")
	fmt.Fprintf(&sb, "<h2 class=\"wp-block-heading\">Excluded</h2>\n")
	fmt.Fprintf(&sb, "<!-- /wp:heading -->\n")

	fmt.Fprintf(&sb, "<!-- wp:table {\"hasFixedLayout\":false,\"align\":\"left\",\"className\":\"is-style-regular\"} -->\n")
	fmt.Fprintf(&sb, "<figure class=\"wp-block-table alignleft is-style-regular\">\n")
	fmt.Fprintf(&sb, "<table><thead><tr><th>Date</th><th>Price</th><th>Categories</th><th>Event &amp; Link</th></tr></thead><tbody>\n")

	for _, event := range allEvents {
		if !event.Include {
			fmt.Fprintf(&sb, "<tr><td>%s</td><td>%s</td><td>", event.Date, event.Price)
			for _, category := range event.Categories {
				// find a color
				color := palette[0]
				for i, cat := range append(cliOptions.Include, cliOptions.Exclude...) {
					if cat == category {
						color = palette[i%len(palette)]
						break
					}
				}
				fmt.Fprintf(&sb, "<mark style=\"background-color:%s\" class=\"has-inline-color has-white-color\"> %s </mark> ", color.Hex(), category)
			}
			fmt.Fprintf(&sb, "</td><td><a href=\"%s\">%s</a></td></tr>\n", event.Link, html.EscapeString(event.Name))
		}
	}

	fmt.Fprintf(&sb, "</tbody></table></figure>\n")
	fmt.Fprintf(&sb, "<!-- /wp:table -->\n")

	return sb.String()
}

func generateTablePress() string {
	var sb strings.Builder

	// sort & display
	//

	fmt.Fprintf(&sb, "{\n")
	fmt.Fprintf(&sb, "  \"name\": \"External events %s to %s\",\n", cliOptions.StartDate, cliOptions.EndDate)
	fmt.Fprintf(&sb, "  \"description\": \"This is a list of events you may be interested in.\",\n")
	fmt.Fprintf(&sb, "  \"data\": [\n")
	fmt.Fprintf(&sb, "    [\n")
	fmt.Fprintf(&sb, "      \"Row (hidden)\",\n")
	fmt.Fprintf(&sb, "      \"Date\",\n")
	fmt.Fprintf(&sb, "      \"Price\",\n")
	fmt.Fprintf(&sb, "      \"Categories\",\n")
	fmt.Fprintf(&sb, "      \"Event & Link\"\n")
	fmt.Fprintf(&sb, "    ],\n")

	for i, event := range allEvents {
		fmt.Fprintf(&sb, "    [\n")
		fmt.Fprintf(&sb, "      \"%d\",\n", i+1)
		fmt.Fprintf(&sb, "      \"%s \",\n", event.Date)
		fmt.Fprintf(&sb, "      \"%s \",\n", event.Price)
		fmt.Fprintf(&sb, "      \"")
		for _, category := range event.Categories {
			// find a color
			color := palette[0]
			for i, cat := range append(cliOptions.Include, cliOptions.Exclude...) {
				if cat == category {
					color = palette[i%len(palette)]
					break
				}
			}
			fmt.Fprintf(&sb, "<mark style=\\\"background-color:%s\\\" class=\\\"has-inline-color has-white-color\\\"> %s </mark> ", color.Hex(), category)
		}
		fmt.Fprintf(&sb, "\",\n")
		fmt.Fprintf(&sb, "      \"<a href=\\\"%s\\\">%s</a>\"\n", event.Link, html.EscapeString(event.Name)) // might need to escape " here
		if i+1 < len(allEvents) {
			fmt.Fprintf(&sb, "    ],\n")
		} else {
			fmt.Fprintf(&sb, "    ]\n")
		}
	}

	fmt.Fprintf(&sb, "  ],\n")
	fmt.Fprintf(&sb, "  \"options\": {\n")
	fmt.Fprintf(&sb, "    \"table_head\": 1,\n")
	fmt.Fprintf(&sb, "    \"table_foot\": 0,\n")
	fmt.Fprintf(&sb, "    \"alternating_row_colors\": true,\n")
	fmt.Fprintf(&sb, "    \"row_hover\": true,\n")
	fmt.Fprintf(&sb, "    \"print_name\": true,\n")
	fmt.Fprintf(&sb, "    \"print_name_position\": \"above\",\n")
	fmt.Fprintf(&sb, "    \"print_description\": true,\n")
	fmt.Fprintf(&sb, "    \"print_description_position\": \"above\",\n")
	fmt.Fprintf(&sb, "    \"extra_css_classes\": \"\",\n")
	fmt.Fprintf(&sb, "    \"use_datatables\": true,\n")
	fmt.Fprintf(&sb, "    \"datatables_sort\": false,\n")
	fmt.Fprintf(&sb, "    \"datatables_filter\": true,\n")
	fmt.Fprintf(&sb, "    \"datatables_paginate\": true,\n")
	fmt.Fprintf(&sb, "    \"datatables_lengthchange\": true,\n")
	fmt.Fprintf(&sb, "    \"datatables_paginate_entries\": 20,\n")
	fmt.Fprintf(&sb, "    \"datatables_info\": true,\n")
	fmt.Fprintf(&sb, "    \"datatables_scrollx\": false,\n")
	fmt.Fprintf(&sb, "    \"datatables_custom_commands\": \"\"\n")
	fmt.Fprintf(&sb, "  },\n")

	fmt.Fprintf(&sb, "  \"visibility\": {\n")
	fmt.Fprintf(&sb, "    \"rows\": [\n")
	fmt.Fprintf(&sb, "      1,\n")
	for i, event := range allEvents {
		if event.Include {
			fmt.Fprintf(&sb, "      1")
		} else {
			fmt.Fprintf(&sb, "      0")
		}
		if i+1 < len(allEvents) {
			fmt.Fprintf(&sb, ",\n")
		} else {
			fmt.Fprintf(&sb, "\n")
		}
	}
	fmt.Fprintf(&sb, "    ],\n")
	fmt.Fprintf(&sb, "    \"columns\": [\n")
	fmt.Fprintf(&sb, "      0,\n")
	fmt.Fprintf(&sb, "      1,\n")
	fmt.Fprintf(&sb, "      1,\n")
	fmt.Fprintf(&sb, "      1,\n")
	fmt.Fprintf(&sb, "      1\n")
	fmt.Fprintf(&sb, "    ]\n")
	fmt.Fprintf(&sb, "  }\n")
	fmt.Fprintf(&sb, "}\n")

	return sb.String()
}

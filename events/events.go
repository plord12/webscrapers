/**

find eventbrite events, output in wordpress format

*/

package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"reflect"
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
	Headless       bool     `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Category       string   `short:"c" long:"category" description:"Category" default:"science-and-tech" env:"CATEGORY"`
	StartDate      string   `short:"s" long:"startdate" description:"Start date (YYYY-MM-DD)" default:"2026-01-01" env:"STARTDATE"`
	EndDate        string   `short:"a" long:"enddate" description:"End date (YYYY-MM-DD)" default:"2030-01-01" env:"ENDDATE"`
	MaxPrice       float64  `short:"p" long:"maxprice" description:"Max price for event (£)" default:"20" env:"PRICE"`
	Nighttime      bool     `short:"n" long:"nighttime" description:"Include nighttime events" env:"NIGHTTIME"`
	Maxpage        int      `short:"m" long:"maxpage" description:"Max page number to fetch" default:"1000" env:"MAXPAGE"`
	Format         string   `short:"f" long:"format" description:"Format - list, table or tablepress" default:"list" choice:"list" choice:"table" choice:"tablepress" env:"FORMAT"`
	Include        []string `short:"i" long:"include" description:"Include - list of categories to include" env:"INCLUDE"`
	Exclude        []string `short:"x" long:"exclude" description:"Exclude - list of categories to exclude" env:"EXCLUDE"`
	Clear          bool     `short:"z" long:"clear" description:"Clear the cache ... eg change in categories" env:"CLEAR"`
	Reclassify     bool     `short:"r" long:"reclassify" description:"Force re-classify" env:"RECLASSIFY"`
	Save           string   `short:"v" long:"save" description:"Filename to save output to" env:"SAVE"`
	Perftest       bool     `short:"t" long:"perftest" description:"Run performance tests only" env:"PERFTEST"`
	CacheAnalyse   bool     `short:"y" long:"cacheanalyse" description:"Run cache analyse tests only" env:"CACHE"`
	OutputExcluded bool     `short:"o" long:"outputexcluded" description:"Output excluded events" env:"OUTPUTEXCLUDED"`
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

type Cache struct {
	Title       string
	Description string
	Price       string
	Date        string
	Categories  []string
}

var eventCache *cache.Cache[Cache]

type record[T interface{}] struct {
	Item       T
	Expiration uint64
}

// constants
//
// Machine Learning
//var mlModelFile = "onnx/model.onnx"
//var mlModel = "knowledgator/gliclass-small-v1.0"

//var mlModelFile = "model.onnx"
//var mlModel = "cnmoro/gliclass-edge-v3.0-onnx"

//
// var mlModel = "KnightsAnalytics/distilbert-base-uncased-finetuned-sst-2-english"
//
// var mlModel = winado/gliclass-base-onnx"
//

var mlModelFile = "onnx/model.onnx"
var mlModel = "MoritzLaurer/deberta-v3-large-zeroshot-v2.0"

// var mlModelFile = "MoritzLaurer/deberta-v3-base-zeroshot-v2.0"
//
// var mlModelFile = "onnx-community/deberta-v3-small"
// const mlModel = "KnightsAnalytics/deberta-v3-base-zeroshot-v1"
// const mlModelFile = "model.onnx"
const mlBackend = "ORT"
const mlMinScore = 0.1

// night time
const nighttimeEndHour = 8
const nighttimeStartHour = 22

const maxCategoriesPerEvent = 3
const maxDescriptionWords = 40

var palette []colorful.Color

var page1 playwright.Page
var page2 playwright.Page

// stats
var eventsFound = 0
var eventsSkippedByDescription = 0
var eventsSkippedByNightTime = 0
var eventsSkippedByPrice = 0
var eventsErrors = 0
var eventBriteIncluded = 0
var greshamIncluded = 0
var rigbIncluded = 0
var yorkIncluded = 0
var uclIncluded = 0

// if found in the cache, must still re-classify since categories have changed
var mustClassify = false

var classificationPipeline *pipelines.ZeroShotClassificationPipeline

// validate arguments
var startDate time.Time
var endDate time.Time

func main() {

	// parse flags
	//
	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}

	if cliOptions.Perftest {
		perftests("ORT", "", 20, "KnightsAnalytics/deberta-v3-base-zeroshot-v1", "model.onnx")
		perftests("ORT", "", 40, "KnightsAnalytics/deberta-v3-base-zeroshot-v1", "model.onnx")
		perftests("ORT", "", 60, "KnightsAnalytics/deberta-v3-base-zeroshot-v1", "model.onnx")
		perftests("ORT", "", 80, "KnightsAnalytics/deberta-v3-base-zeroshot-v1", "model.onnx")
		// mac 2.9s arm6 10.7s
		perftests("ORT", "", 20, mlModel, mlModelFile)
		perftests("ORT", "", 40, mlModel, mlModelFile)
		perftests("ORT", "", 60, mlModel, mlModelFile)
		perftests("ORT", "", 80, mlModel, mlModelFile)

		// mac 12.6s, arm6 1m6s
		//perftests("XLA", "", maxDescriptionWords)
		// mac 10.7s, arm6 23.9s
		//perftests("ORT", "XNNPACK", maxDescriptionWords)
		// crashes
		// perftests("ORT", "CoreML", 20, mlModel, mlModelFile)
		// doesn't work
		//perftests("ORT", "ACL")
		// never ends
		// perftests("", "", maxDescriptionWords, mlModel, mlModelFile)
		os.Exit(0)
	}

	if cliOptions.CacheAnalyse {
		cacheAnalyse()
		os.Exit(0)
	}

	getExchangeRates()

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

	// disk cache ... perhaps this should be the same as Event ?
	//

	eventCache = cache.New[Cache]("events", cache.CacheOptionPersistent).WithExpiration(7 * 24 * time.Hour)
	if cliOptions.Clear {
		eventCache.Clear()
	} else {
		lastRuncategories, err := eventCache.Get("all categories")
		if err == nil && reflect.DeepEqual(lastRuncategories.Categories, append(cliOptions.Include, cliOptions.Exclude...)) {
			// all good
		} else {
			mustClassify = true
			fmt.Fprintf(os.Stderr, "Categories have changed, have to re-run classifications\n")
		}
	}
	eventCache.Set(Cache{Categories: append(cliOptions.Include, cliOptions.Exclude...)}, "all categories")

	// FIX THIS - add https://www.linnean.org/meetings-and-events Linnean Society two or three
	// FIX THIS - add https://www.bcs.org/events-calendar/ BCS (the Chartered Institute for IT) several hybrid or webinar items each month. Booked through Eventbrite. But not all appear under science and tech
	// FIX THIS - add https://kipac.stanford.edu/events/upcoming-events KIPAC (Kavli Institute for particle Astrophysics and cosmology) Stanford University several items each month

	// machine learning classification
	//
	var session *hugot.Session
	switch mlBackend {
	case "XLA":
		session, err = hugot.NewXLASession()
	case "ORT":
		session, err = hugot.NewORTSession()
	default:
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
	classificationPipeline, err = hugot.NewPipeline(session, config)
	if err != nil {
		panic(fmt.Sprintf("could not create pipeline: %v", err))
	}

	// setup web browser windows
	//
	page1 = utils.StartCamoufox(cliOptions.Headless)
	defer utils.Finish(page1)

	newContext, err := page1.Context().Browser().NewContext()
	if err != nil {
		panic(fmt.Sprintf("could not open new page: %v", err))
	}
	page2, err = newContext.NewPage()
	if err != nil {
		panic(fmt.Sprintf("could not open new page: %v", err))
	}
	defer utils.Finish(page2)

	// get events
	//
	eventbrite()
	gresham()
	rigb()
	york()
	ucl()

	// summary report
	//
	fmt.Printf("events has been run with the following options :\n")
	fmt.Printf("	Headless=%v\n", cliOptions.Headless)
	fmt.Printf("	Category=%s\n", cliOptions.Category)
	fmt.Printf("	Date=%s to %s\n", cliOptions.StartDate, cliOptions.EndDate)
	fmt.Printf("	Nighttime=%v\n", cliOptions.Nighttime)
	fmt.Printf("	Maxpage=%d\n", cliOptions.Maxpage)
	fmt.Printf("	Format=%s\n", cliOptions.Format)
	fmt.Printf("	Max price=£%.2f\n", cliOptions.MaxPrice)
	fmt.Printf("	Include=%s\n", strings.Join(cliOptions.Include, ","))
	fmt.Printf("	Exclude=%s\n", strings.Join(cliOptions.Exclude, ","))
	fmt.Printf("	Output excluded=%v (either via seperate tables or hidden rows)\n", cliOptions.OutputExcluded)
	fmt.Printf("	Machine learning model %s with %s backend\n", mlModel, mlBackend)
	fmt.Printf("\n")
	fmt.Printf("There were %d events found.  Of which :\n", eventsFound)
	fmt.Printf("	%d were skipped due to excluded categories match\n", eventsSkippedByDescription)
	fmt.Printf("	%d were skipped due to nighttime\n", eventsSkippedByNightTime)
	fmt.Printf("	%d were skipped due to high price\n", eventsSkippedByPrice)
	fmt.Printf("	%d were included from eventbrite\n", eventBriteIncluded)
	fmt.Printf("	%d were included from gresham\n", greshamIncluded)
	fmt.Printf("	%d were included from royal institution\n", rigbIncluded)
	fmt.Printf("	%d were included from university of york\n", yorkIncluded)
	fmt.Printf("	%d were included from university college london\n", uclIncluded)
	fmt.Printf("	%d errors\n", eventsErrors)
	fmt.Printf("\n")

	// sort
	//
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Sort < allEvents[j].Sort
	})

	// colour pallet
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

	fmt.Printf("Attached json files can be imported into wordpress tablepress - simply dragging the file from email to\n")
	fmt.Printf("wordpress tablepress import field should work.\n")
	fmt.Printf("\n")

	fmt.Printf("Attached wordpress html files can be cut&pasted onto your page.  Switch to the `Code editor` (top right menu),\n")
	fmt.Printf("paste then switch back to `Visual editor`.  A cut&paste to the mailpoet editor should also work.\n")
	fmt.Printf("\n")

	// and generate report
	//
	report := ""
	switch cliOptions.Format {
	case "list":
		report = generateList()
	case "table":
		report = generateTable()
	default:
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

func generateList() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "<!-- wp:heading -->\n")
	fmt.Fprintf(&sb, "<h2 class=\"wp-block-heading\">%s to %s</h2>\n", cliOptions.StartDate, cliOptions.EndDate)
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

	if !cliOptions.OutputExcluded {
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
	}
	return sb.String()
}

func generateTable() string {
	var sb strings.Builder

	// sort & display
	//

	fmt.Printf("<!-- wp:heading -->\n")
	fmt.Fprintf(&sb, "<h2 class=\"wp-block-heading\">%s to %s</h2>\n", cliOptions.StartDate, cliOptions.EndDate)
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

	if !cliOptions.OutputExcluded {
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
	}

	return sb.String()
}

func generateTablePress() string {

	type tablePressOptions struct {
		TableHead                  int    `json:"table_head"`
		TableFoot                  int    `json:"table_foot"`
		AlternatingRowColors       bool   `json:"alternating_row_colors"`
		RowHover                   bool   `json:"row_hover"`
		PrintName                  bool   `json:"print_name"`
		PrintNamePositition        string `json:"print_name_position"`
		PrintDescription           bool   `json:"print_description"`
		PrintDescriptionPositition string `json:"print_description_position"`
		ExtraCssClasses            string `json:"extra_css_classes"`
		UseDataTables              bool   `json:"use_datatables"`
		DataTablesSort             bool   `json:"datatables_sort"`
		DataTablesFilter           bool   `json:"datatables_filter"`
		DataTablesPaginate         bool   `json:"datatables_paginate"`
		DataTablesLengthChange     bool   `json:"datatables_lengthchange"`
		DataTablesPaginateEntries  int    `json:"datatables_paginate_entries"`
		DataTablesInfo             bool   `json:"datatables_info"`
		DataTablesScrollX          bool   `json:"datatables_scrollx"`
		DataTablesCustomCommand    string `json:"datatables_custom_commands"`
	}
	type visibility struct {
		Rows    []int `json:"rows"`
		Columns []int `json:"columns"`
	}
	type tablePress struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Data        [][]string        `json:"data"`
		Options     tablePressOptions `json:"options"`
		Visibility  visibility        `json:"visibility"`
	}

	var tablePressStruct tablePress

	tablePressStruct.Options.TableHead = 1
	tablePressStruct.Options.TableFoot = 0
	tablePressStruct.Options.AlternatingRowColors = true
	tablePressStruct.Options.RowHover = true
	tablePressStruct.Options.PrintName = true
	tablePressStruct.Options.PrintNamePositition = "above"
	tablePressStruct.Options.PrintDescription = true
	tablePressStruct.Options.PrintDescriptionPositition = "above"
	tablePressStruct.Options.UseDataTables = true
	tablePressStruct.Options.DataTablesSort = false
	tablePressStruct.Options.DataTablesFilter = true
	tablePressStruct.Options.DataTablesPaginate = true
	tablePressStruct.Options.DataTablesLengthChange = true
	tablePressStruct.Options.DataTablesPaginateEntries = 20
	tablePressStruct.Options.DataTablesInfo = true
	tablePressStruct.Options.DataTablesScrollX = false

	tablePressStruct.Name = fmt.Sprintf("External events %s to %s", cliOptions.StartDate, cliOptions.EndDate)
	tablePressStruct.Description = "This is a list of events you may be interested in."

	tablePressStruct.Data = append(tablePressStruct.Data, []string{"Row (hidden)", "Date", "Price", "Categories", "Event & Link"})
	tablePressStruct.Visibility.Rows = append(tablePressStruct.Visibility.Rows, 1)
	tablePressStruct.Visibility.Columns = []int{0, 1, 1, 1, 1}

	for i, event := range allEvents {
		if cliOptions.OutputExcluded && !event.Include {
			// skip
		} else {
			categories := ""
			for _, category := range event.Categories {
				// find a color
				color := palette[0]
				for i, cat := range append(cliOptions.Include, cliOptions.Exclude...) {
					if cat == category {
						color = palette[i%len(palette)]
						break
					}
				}
				categories = categories + fmt.Sprintf("<mark style=\"background-color:%s\" class=\"has-inline-color has-white-color\"> %s </mark> ", color.Hex(), category)
			}
			description := fmt.Sprintf("<a href=\"%s\">%s</a>\n", event.Link, html.EscapeString(event.Name))
			tablePressStruct.Data = append(tablePressStruct.Data, []string{strconv.Itoa(i + 1), event.Date, event.Price, categories, description})
			if event.Include {
				tablePressStruct.Visibility.Rows = append(tablePressStruct.Visibility.Rows, 1)
			} else {
				tablePressStruct.Visibility.Rows = append(tablePressStruct.Visibility.Rows, 0)
			}
		}
	}

	tablePressJson, _ := json.Marshal(tablePressStruct)

	return string(tablePressJson)
}

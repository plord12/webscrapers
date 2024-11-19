/**

Get fund value

*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	stealth "github.com/jonfriesen/playwright-go-stealth"
	"github.com/playwright-community/playwright-go"
)

var page playwright.Page
var pw *playwright.Playwright

func finish() {
	page.Close()

	// on error, save video if we can
	r := recover()
	if r != nil {
		log.Println("Failure:", r)
		path, err := page.Video().Path()
		if err == nil {
			log.Printf("Final screen video saved at %s\n", path)
		} else {
			log.Printf("Failed to save final video: %v\n", err)
		}
	} else {
		page.Video().Delete()
	}

	pw.Stop()
}

func main() {

	// defaults from environment
	//
	defaultHeadless := true
	defaultFund := ""

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envFund := os.Getenv("FUND"); envFund != "" {
		defaultFund = envFund
	}

	// arguments
	//
	headless := flag.Bool("headless", defaultHeadless, "Headless mode")

	fund := flag.String("fund", defaultFund, "Fund name")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Retrive fund value via web scraping")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $FUND - Fund name")
	}

	// parse flags
	//
	flag.Parse()

	// FIX THIS - validate

	// setup
	//
	err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
	if err != nil {
		panic(fmt.Sprintf("could not install playwright: %v", err))
	}
	pw, err = playwright.Run()
	if err != nil {
		panic(fmt.Sprintf("could not launch playwright: %v", err))
	}
	defer finish()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(*headless)})
	if err != nil {
		panic(fmt.Sprintf("could not launch Chromium: %v", err))
	}
	page, err = browser.NewPage(playwright.BrowserNewPageOptions{RecordVideo: &playwright.RecordVideo{Dir: "videos/"}})
	if err != nil {
		panic(fmt.Sprintf("could not create page: %v", err))
	}
	page.SetDefaultTimeout(10000)

	// Inject stealth script
	//
	err = stealth.Inject(page)
	if err != nil {
		panic(fmt.Sprintf("could not inject stealth script: %v", err))
	}

	// main page & login
	//
	log.Printf("Starting fund\n")
	_, err = page.Goto("https://markets.ft.com/data/funds/tearsheet/summary?s="+*fund, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	// get value
	//

	// <span class="mod-ui-data-list__value">11.89</span>
	value, err := page.Locator("[class=mod-ui-data-list__value]").First().TextContent()
	if err != nil {
		panic(fmt.Sprintf("failed to get balance: %v", err))
	}
	log.Println("value=" + value)
	fmt.Println(value)

	bufio.NewWriter(os.Stdout).Flush()

}

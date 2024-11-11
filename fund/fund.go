/**

Get fund value

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	stealth "github.com/jonfriesen/playwright-go-stealth"
	"github.com/playwright-community/playwright-go"
)

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
		log.Fatalf("could not install playwright: %v", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not launch playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(*headless)})
	if err != nil {
		pw.Stop()
		log.Fatalf("could not launch Chromium: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not create page: %v", err)
	}
	// Inject stealth script
	//
	err = stealth.Inject(page)
	if err != nil {
		log.Fatalf("could not inject stealth script: %v", err)
	}

	// main page & login
	//
	log.Printf("Starting chromium\n")
	_, err = page.Goto("https://markets.ft.com/data/funds/tearsheet/summary?s="+*fund, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not goto url: %v", err)
	}

	// get value
	//

	// <span class="mod-ui-data-list__value">11.89</span>
	value, err := page.Locator("[class=mod-ui-data-list__value]").First().TextContent()
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("failed to get balance: %v", err)
	}
	log.Println("value=" + value)
	fmt.Println(value)

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

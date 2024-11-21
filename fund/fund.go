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

	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
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
	page := utils.StartChromium(headless)
	defer utils.Finish(page)

	page.SetDefaultTimeout(10000)

	// main page & login
	//
	log.Printf("Starting fund\n")
	_, err := page.Goto("https://markets.ft.com/data/funds/tearsheet/summary?s="+*fund, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
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

/**

Get fund value

*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Fund     string `short:"f" long:"fund" description:"Fund name" env:"FUND" required:"true"`
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

	page.SetDefaultTimeout(10000)

	// main page & login
	//
	log.Printf("Starting fund\n")
	_, err = page.Goto("https://markets.ft.com/data/funds/tearsheet/summary?s="+*&options.Fund, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
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

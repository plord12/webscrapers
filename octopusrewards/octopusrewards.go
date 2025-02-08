/**

report on octopus rewards

*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username string `short:"u" long:"username" description:"Octopus username" env:"OCTOPUS_USERNAME" required:"true"`
	Password string `short:"p" long:"password" description:"Octopus password" env:"OCTOPUS_PASSWORD" required:"true"`
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

	// main page & login
	//
	_, err = page.Goto("https://octopus.energy/login/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	err = page.Locator("#id_username").Fill(options.Username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	err = page.Locator("#id_password").Fill(options.Password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}
	err = page.Locator(".button").Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	err = page.GetByText("Claim rewards").First().Click()
	if err != nil {
		panic("Could not find first Claim rewards")
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateLoad})

	err = page.GetByText("Explore rewards").First().Click()
	if err != nil {
		panic("Could not find first Explore rewards")
	}
	page.GetByText("All Offers")
	time.Sleep(1 * time.Second)
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateLoad})

	offers, err := page.Locator("h3").All()
	if err != nil {
		panic("Could not find offer descriptions")
	}
	for _, offer := range offers {
		fmt.Println(offer.First().TextContent())
	}

	bufio.NewWriter(os.Stdout).Flush()
}

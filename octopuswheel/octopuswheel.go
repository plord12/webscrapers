/**

sping octopus wheel

*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

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
	log.Printf("Starting login\n")
	_, err = page.Goto("https://octopus.energy/login/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	log.Printf("Logging in\n")
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
	err = page.Locator(".jAWbYk").Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}
	split := strings.Split(page.URL(), "/")
	var account string
	if len(split) > 6 {
		account = split[6]
	} else {
		panic("could not get account")
	}

	page.SetDefaultTimeout(5000)

	log.Println("Spinning electricity for " + account)
	// electricity
	//
	_, err = page.Goto("https://octopus.energy/dashboard/new/accounts/"+account+"/wheel-of-fortune/electricity", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}
	page.Locator(".wheel").Click()

	log.Println("Spinning gas for " + account)
	// electricity
	//
	_, err = page.Goto("https://octopus.energy/dashboard/new/accounts/"+account+"/wheel-of-fortune/gas", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}
	page.Locator(".wheel").Click()

	bufio.NewWriter(os.Stdout).Flush()
}

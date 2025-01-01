/**

sping octopus wheel

*/

package main

import (
	"bufio"
	"fmt"
	"log"
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

	page.SetDefaultTimeout(5000)

	log.Printf("Looking for first wheel of fortune\n")
	err = page.GetByText("Spin the Wheel of Fortune").First().Click()
	if err != nil {
		log.Printf("Could not find first wheel of fortune\n")
	} else {
		page.Locator(".wheel").Click()
		time.Sleep(5 * time.Second)
		log.Printf("Done\n")
		page.GoBack()
	}

	log.Printf("Looking for last wheel of fortune\n")
	err = page.GetByText("Spin the Wheel of Fortune").Last().Click()
	if err != nil {
		log.Printf("Could not find last wheel of fortune\n")
	} else {
		page.Locator(".wheel").Click()
		time.Sleep(5 * time.Second)
		log.Printf("Done\n")
		page.GoBack()
	}

	bufio.NewWriter(os.Stdout).Flush()
}

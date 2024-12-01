/**

Update meoneyhub balance

*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless bool     `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username string   `short:"u" long:"username" description:"Moneyhub username" env:"MONEYHUB_USERNAME" required:"true"`
	Password string   `short:"p" long:"password" description:"Moneyhub password" env:"MONEYHUB_PASSWORD" required:"true"`
	Accounts []string `short:"a" long:"account" description:"Moneyhub account(s)" env:"MONEYHUB_ACCOUNT" env-delim:"," required:"true"`
	Balances []string `short:"b" long:"balance" description:"Moneyhub balance(s)" env:"MONEYHUB_BALANCE" env-delim:"," required:"true"`
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

	if len(options.Accounts) < 1 || len(options.Balances) < 1 || len(options.Accounts) != len(options.Balances) {
		panic("accounts and balances do not match")
	}

	// FIX THIS - validate

	// setup
	//
	page := utils.StartChromium(options.Headless)
	defer utils.Finish(page)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err = page.Goto("https://client.moneyhub.co.uk", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	log.Printf("Logging in\n")
	// <input name="email" id="email" autocomplete="username" data-aid="field-email" type="email" class="sc-eNQAEJ hgPDnc" value="">
	err = page.Locator("#email").Fill(options.Username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	// <input name="password" id="password" data-aid="field-password" type="password" minlength="10" autocomplete="current-password" class="sc-hSdWYo kBIXYI" value="">
	err = page.Locator("#password").Fill(options.Password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}
	// <span class="sc-bxivhb sc-ifAKCX byYfdZ">Log in</span>
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Log in"}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	var failed bool = false

	for i := 0; i < len(options.Accounts); i++ {

		// goto assets & update
		//
		_, err = page.Goto("https://client.moneyhub.co.uk/#assets", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			log.Printf("could not goto assets: %v", err)
			failed = true
			continue
		}
		// occational "Stay Connected" pop-up
		page.GetByText("Stay connected", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(500.0)})

		// <div data-aid="ListItemTitle" class="sc-bxivhb list-item-title__Title-sc-uq1r70-0 bOSooI">Peter Moneyfarm ISA [ manual ]</div>
		err = page.GetByText(options.Accounts[i], playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Delay: playwright.Float(500.0)})
		if err != nil {
			log.Printf("could not goto asset: %s %v", options.Accounts[i], err)
			failed = true
			continue
		}
		// <button label="appChrome.edit" data-aid="nav-bar-edit" aria-label="Edit Account" class="button__Button-sc-182rbpd-0 czyaZa"><div height="32px" width="32px" style="pointer-events: none;" aria-hidden="true"><div>...
		err = page.Locator("[label=\"appChrome.edit\"]").Click(playwright.LocatorClickOptions{Delay: playwright.Float(500.0)})
		if err != nil {
			log.Printf("could not edit asset: %s %v", options.Accounts[i], err)
			failed = true
			continue
		}
		//<span class="sc-bxivhb sc-ifAKCX byYfdZ">Update valuation</span>
		//<span class="sc-bxivhb sc-ifAKCX byYfdZ">Update balance</span>
		err = page.GetByText(regexp.MustCompile("^Update ")).Click(playwright.LocatorClickOptions{Delay: playwright.Float(500.0)})
		if err != nil {
			log.Printf("could not update asset: %s %v", options.Accounts[i], err)
			failed = true
			continue
		}
		// <input name="balance" id="balance" type="text" inputmode="decimal" pattern="[0-9]*.?[0-9]*" autocomplete="off" class="sc-cSHVUG jVBxUm" value="92276.76">
		err = page.Locator("#balance").Clear()
		if err != nil {
			log.Printf("could not clear balance: %s %v", options.Accounts[i], err)
			failed = true
			continue
		}
		err = page.Locator("#balance").Fill(options.Balances[i])
		if err != nil {
			log.Printf("could not update balance: %s %v", options.Accounts[i], err)
			failed = true
			continue
		}
		err = page.GetByText("Save", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
		if err != nil {
			log.Printf("could not save balance: %s %v", options.Accounts[i], err)
			failed = true
			continue
		} else {
			log.Println("Account " + options.Accounts[i] + " updated to " + options.Balances[i])
		}

		page.Reload()

	}

	if failed {
		panic("One or more accounts couldn't be updated")
	}

	bufio.NewWriter(os.Stdout).Flush()
}

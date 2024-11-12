/**

Update meoneyhub balance

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	stealth "github.com/jonfriesen/playwright-go-stealth"
	"github.com/playwright-community/playwright-go"
)

func main() {

	// defaults from environment
	//
	defaultHeadless := true
	defaultUsername := ""
	defaultPassword := ""
	defaultAccount := ""
	defaultBalance := 0.0

	var err error

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envUsername := os.Getenv("MONEYHUB_USERNAME"); envUsername != "" {
		defaultUsername = envUsername
	}
	if envPassword := os.Getenv("MONEYHUB_PASSWORD"); envPassword != "" {
		defaultPassword = envPassword
	}
	if envAccount := os.Getenv("MONEYHUB_ACCOUNT"); envAccount != "" {
		defaultAccount = envAccount
	}
	if envBalance := os.Getenv("MONEYHUB_BALANCE"); envBalance != "" {
		defaultBalance, err = strconv.ParseFloat(envBalance, 32)
		if err != nil {
			log.Fatalf("could not conert balance: %v", err)
		}
	}

	// arguments
	//
	headless := flag.Bool("headless", defaultHeadless, "Headless mode")

	username := flag.String("username", defaultUsername, "Moneyhub username")
	password := flag.String("password", defaultPassword, "Moneyhub password")
	account := flag.String("account", defaultAccount, "Moneyhub account")
	balance := flag.Float64("balance", defaultBalance, "Moneyhub balance for the account")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Update Moneyhub balance via web scraping")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $MONEYHUB_USERNAME - Moneyhub username")
		fmt.Println("  $MONEYHUB_PASSWORD - Moneyhub password")
		fmt.Println("  $MONEYHUB_ACCOUNT - Moneyhub account")
		fmt.Println("  $MONEYHUB_BALANCE - Moneyhub balance for the account")
	}

	// parse flags
	//
	flag.Parse()

	// FIX THIS - validate

	// setup
	//
	err = playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
	if err != nil {
		log.Fatalf("could not install playwright: %v", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not launch playwright: %v", err)
	}
	defer pw.Stop()
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(*headless)})
	if err != nil {
		log.Fatalf("could not launch Chromium: %v", err)
	}
	defer browser.Close()
	page, err := browser.NewPage()
	if err != nil {
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
	log.Printf("Starting login\n")
	_, err = page.Goto("https://client.moneyhub.co.uk", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		log.Fatalf("could not goto url: %v", err)
	}

	log.Printf("Logging in\n")
	// <input name="email" id="email" autocomplete="username" data-aid="field-email" type="email" class="sc-eNQAEJ hgPDnc" value="">
	err = page.Locator("#email").Fill(*username)
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}
	// <input name="password" id="password" data-aid="field-password" type="password" minlength="10" autocomplete="current-password" class="sc-hSdWYo kBIXYI" value="">
	err = page.Locator("#password").Fill(*password)
	if err != nil {
		log.Fatalf("could not get password: %v", err)
	}
	// <span class="sc-bxivhb sc-ifAKCX byYfdZ">Log in</span>
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Log in"}).Click()
	if err != nil {
		log.Fatalf("could not click: %v", err)
	}

	// goto assets & update
	//
	_, err = page.Goto("https://client.moneyhub.co.uk/#assets", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		log.Fatalf("could not goto assets: %v", err)
	}
	// <div data-aid="ListItemTitle" class="sc-bxivhb list-item-title__Title-sc-uq1r70-0 bOSooI">Peter Moneyfarm ISA [ manual ]</div>
	var delay = 500.0
	err = page.GetByText(*account, playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Delay: &delay})
	if err != nil {
		log.Fatalf("could not goto asset: %v", err)
	}
	// <button label="appChrome.edit" data-aid="nav-bar-edit" aria-label="Edit Account" class="button__Button-sc-182rbpd-0 czyaZa"><div height="32px" width="32px" style="pointer-events: none;" aria-hidden="true"><div>...
	err = page.Locator("[label=\"appChrome.edit\"]").Click()
	if err != nil {
		log.Fatalf("could not edit asset: %v", err)
	}
	//<span class="sc-bxivhb sc-ifAKCX byYfdZ">Update valuation</span>
	//<span class="sc-bxivhb sc-ifAKCX byYfdZ">Update balance</span>
	err = page.GetByText(regexp.MustCompile("^Update ")).Click()
	if err != nil {
		log.Fatalf("could not update asset: %v", err)
	}
	// <input name="balance" id="balance" type="text" inputmode="decimal" pattern="[0-9]*.?[0-9]*" autocomplete="off" class="sc-cSHVUG jVBxUm" value="92276.76">
	err = page.Locator("#balance").Clear()
	if err != nil {
		log.Fatalf("could not clear balance: %v", err)
	}
	err = page.Locator("#balance").Fill(fmt.Sprintf("%0.2f", *balance))
	if err != nil {
		log.Fatalf("could not update balance: %v", err)
	}
	err = page.GetByText("Save", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
	if err != nil {
		log.Fatalf("could not save balance: %v", err)
	}

	log.Println("Account " + *account + " updated to " + fmt.Sprintf("%0.2f", *balance))
}

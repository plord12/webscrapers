/**

Get aviva my money balance

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	stealth "github.com/jonfriesen/playwright-go-stealth"
	"github.com/playwright-community/playwright-go"
)

func main() {

	// defaults from environment
	//
	defaultHeadless := true
	defaultUsername := ""
	defaultPassword := ""
	defaultWord := ""

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envUsername := os.Getenv("AVIVAMYMONEY_USERNAME"); envUsername != "" {
		defaultUsername = envUsername
	}
	if envPassword := os.Getenv("AVIVAMYMONEY_PASSWORD"); envPassword != "" {
		defaultPassword = envPassword
	}
	if envWord := os.Getenv("AVIVAMYMONEY_WORD"); envWord != "" {
		defaultWord = envWord
	}

	// arguments
	//
	headless := flag.Bool("headless", defaultHeadless, "Headless mode")

	username := flag.String("username", defaultUsername, "Aviva my money username")
	password := flag.String("password", defaultPassword, "Aviva my money password")
	word := flag.String("word", defaultWord, "Aviva my money memorable word")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Retrive Aviva my money balance via web scraping")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $AVIVAMYMONEY_USERNAME - Aviva my money username")
		fmt.Println("  $AVIVAMYMONEY_PASSWORD - Aviva my money password")
		fmt.Println("  $AVIVAMYMONEY_WORD - Aviva my money memorable word")
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
	_, err = page.Goto("https://www.avivamymoney.co.uk/Login", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not goto url: %v", err)
	}

	// dismiss pop-up
	//
	page.GetByText("Accept all cookies", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()

	log.Printf("Logging in\n")
	// <input autocomplete="off" id="Username" maxlength="50" name="Username" tabindex="1" type="text" value="">
	err = page.Locator("#Username").Fill(*username)
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not get username: %v", err)
	}
	// <input autocomplete="off" id="Username" maxlength="50" name="Username" tabindex="1" type="text" value="">
	err = page.Locator("#Password").Fill(*password)
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not get password: %v", err)
	}
	// <a href="#" "="" name="undefined" class="btn-primary full-width">Log in</a>
	err = page.GetByText("Log in", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not click: %v", err)
	}

	// memorable word
	//
	page.Locator("#FirstLetter")

	// <input id="FirstElement_Index" name="FirstElement.Index" type="hidden" value="3">
	firstIndex, err := page.Locator("#FirstElement_Index").GetAttribute("value")
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not get first element index: %v", err)
	}
	firstIndexInt, _ := strconv.Atoi(firstIndex)
	page.Locator("#FirstLetter").SelectOption(playwright.SelectOptionValues{ValuesOrLabels: &[]string{(*word)[firstIndexInt-1 : firstIndexInt]}})

	secondIndex, err := page.Locator("#SecondElement_Index").GetAttribute("value")
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not get second element index: %v", err)
	}
	secondIndexInt, _ := strconv.Atoi(secondIndex)
	page.Locator("#SecondLetter").SelectOption(playwright.SelectOptionValues{ValuesOrLabels: &[]string{(*word)[secondIndexInt-1 : secondIndexInt]}})

	thirdIndex, err := page.Locator("#ThirdElement_Index").GetAttribute("value")
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not get third element index: %v", err)
	}
	thirdIndexInt, _ := strconv.Atoi(thirdIndex)
	page.Locator("#ThirdLetter").SelectOption(playwright.SelectOptionValues{ValuesOrLabels: &[]string{(*word)[thirdIndexInt-1 : thirdIndexInt]}})

	err = page.GetByText("Next", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not click next: %v", err)
	}

	// get balance
	//
	// <p class="vspace-reset text-size-42">£22,332.98</p>
	balance, err := page.Locator(".vspace-reset.text-size-42").TextContent()
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("failed to get balance: %v", err)
	}
	log.Println("balance=" + balance)
	fmt.Println(strings.NewReplacer("£", "", ",", "").Replace(balance))

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

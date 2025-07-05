/**

Get aviva my money balance

*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username string `short:"u" long:"username" description:"Aviva my money username" env:"AVIVAMYMONEY_USERNAME" required:"true"`
	Password string `short:"p" long:"password" description:"Aviva my money password" env:"AVIVAMYMONEY_PASSWORD" required:"true"`
	Word     string `short:"w" long:"word" description:"Aviva my money memorable word" env:"AVIVAMYMONEY_WORD" required:"true"`
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
	page := utils.StartCamoufox(options.Headless)
	defer utils.Finish(page)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err = page.Goto("https://www.avivamymoney.co.uk/Login", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	// dismiss pop-up
	//
	page.GetByText("Essential cookies only", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()

	log.Printf("Logging in\n")
	// <input autocomplete="off" id="Username" maxlength="50" name="Username" tabindex="1" type="text" value="">
	err = page.Locator("#Username").Fill(options.Username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	// <input autocomplete="off" id="Username" maxlength="50" name="Username" tabindex="1" type="text" value="">
	err = page.Locator("#Password").Fill(options.Password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}
	// <a href="#" "="" name="undefined" class="btn-primary full-width">Log in</a>
	err = page.GetByText("Log in", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// memorable word
	//
	page.Locator("#FirstLetter")

	// <input id="FirstElement_Index" name="FirstElement.Index" type="hidden" value="3">
	firstIndex, err := page.Locator("#FirstElement_Index").GetAttribute("value")
	if err != nil {
		panic(fmt.Sprintf("could not get first element index: %v", err))
	}
	firstIndexInt, _ := strconv.Atoi(firstIndex)
	page.Locator("#FirstLetter").SelectOption(playwright.SelectOptionValues{ValuesOrLabels: &[]string{(options.Word)[firstIndexInt-1 : firstIndexInt]}})

	secondIndex, err := page.Locator("#SecondElement_Index").GetAttribute("value")
	if err != nil {
		panic(fmt.Sprintf("could not get second element index: %v", err))
	}
	secondIndexInt, _ := strconv.Atoi(secondIndex)
	page.Locator("#SecondLetter").SelectOption(playwright.SelectOptionValues{ValuesOrLabels: &[]string{(options.Word)[secondIndexInt-1 : secondIndexInt]}})

	thirdIndex, err := page.Locator("#ThirdElement_Index").GetAttribute("value")
	if err != nil {
		panic(fmt.Sprintf("could not get third element index: %v", err))
	}
	thirdIndexInt, _ := strconv.Atoi(thirdIndex)
	page.Locator("#ThirdLetter").SelectOption(playwright.SelectOptionValues{ValuesOrLabels: &[]string{(options.Word)[thirdIndexInt-1 : thirdIndexInt]}})

	err = page.GetByText("Next", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click next: %v", err))
	}

	// get balance
	//
	// <p class="vspace-reset text-size-42">£22,332.98</p>
	balance, err := page.Locator(".vspace-reset.text-size-42").TextContent()
	if err != nil {
		panic(fmt.Sprintf("failed to get balance: %v", err))
	}
	log.Println("balance=" + balance)
	fmt.Println(strings.NewReplacer("£", "", ",", "").Replace(balance))

	bufio.NewWriter(os.Stdout).Flush()
}

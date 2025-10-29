/**

Get aviva my workplace balance

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
	Username string `short:"u" long:"username" description:"Aviva my workplace username" env:"AVIVAMYWORKPLACE_USERNAME" required:"true"`
	Password string `short:"p" long:"password" description:"Aviva my workplace password" env:"AVIVAMYWORKPLACE_PASSWORD" required:"true"`
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
	_, err = page.Goto("https://zzz.myworkplace.aviva.co.uk/MyAccount/login", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	// dismiss pop-up
	//
	page.GetByText("Essential cookies only", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()

	log.Printf("Logging in\n")
	// <input aria-required="True" autocomplete="off" class="a-textbox" data-qa-textbox="username" data-val="true" data-val-required="Please enter your username" id="username" maxlength="50" name="username" type="text" value="">
	err = page.Locator("#username").Fill(options.Username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	// <input aria-required="True" autocomplete="off" class="a-textbox" data-qa-textbox="password" data-val="true" data-val-required="Please enter your password" id="password" maxlength="300" name="password" type="password">
	err = page.Locator("#password").Fill(options.Password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}
	// <input id="loginButton" name="loginButton" class="a-button a-button--primary dd-data-link" data-dd-group="myAvivaLogin" data-dd-loc="login" data-dd-link="login" type="submit" value="Log in" data-qa-button="submitForm">
	err = page.GetByText("Log in", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// click through
	// <a class="a-button a-button--tertiary a-button--tertiary-shallow dd-data-link" data-qa-navbutton="viewArrangement" data-dd-group="myavivaHomePage-workplace" data-dd-loc="employer_card" data-dd-link="TheUniversityofReading_homepage" href="/Workplace/Transfer/A5013"><span class="a-button__inner">View <span class="a-button__inner" data-di-mask="">The University of Reading</span></span></a>
	err = page.GetByText("View", playwright.PageGetByTextOptions{Exact: playwright.Bool(false)}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// get balance
	//
	// <p class="m-data-group-item__data" data-qa-policydetail="totalFundValue" data-di-mask="">£27,138.05</p>
	balance, err := page.Locator("[data-qa-policydetail=totalFundValue]").TextContent()
	if err != nil {
		panic(fmt.Sprintf("failed to get balance: %v", err))
	}
	log.Println("balance=" + balance)
	fmt.Println(strings.NewReplacer("£", "", ",", "").Replace(balance))

	bufio.NewWriter(os.Stdout).Flush()
}

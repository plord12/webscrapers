/**

Get aviva balance

*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/plord12/webscrapers/utils"

	"github.com/playwright-community/playwright-go"
)

type Options struct {
	Headless        bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username        string `short:"u" long:"username" description:"Aviva username" env:"AVIVA_USERNAME" required:"true"`
	Password        string `short:"p" long:"password" description:"Aviva password" env:"AVIVA_PASSWORD" required:"true"`
	Otppath         string `short:"o" long:"otppath" description:"Path to file containing one time password message" default:"otp/aviva" env:"OTP_PATH"`
	Otpcommand      string `short:"c" long:"otpcommand" description:"Command to get one time password" env:"OTP_COMMAND"`
	Otpcleancommand string `short:"l" long:"otpcleancommand" description:"Command to clean previous one time password" env:"OTP_CLEANCOMMAND"`
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

	// clean from any previous run
	//
	utils.CleanOTP(options.Otpcleancommand, options.Otpcleancommand)

	// setup
	//
	page := utils.StartCamoufox(options.Headless)
	defer utils.Finish(page)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err = page.Goto("https://www.direct.aviva.co.uk/MyAccount/login", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	// dismiss pop-up
	//
	// <button id="onetrust-accept-btn-handler">Accept all cookies</button>
	page.GetByText("Accept all cookies", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

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
	err = page.Locator("#loginButton").Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// dismiss pop-up
	//
	// <button id="onetrust-accept-btn-handler">Accept all cookies</button>
	page.GetByText("Accept all cookies", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

	// attempt to fetch one time password if needed
	//
	utils.FetchOTP(options.Otpcommand)

	// check/poll if otp/aviva exists ... could be via the above command or pushed here elsewhere
	//
	otp := utils.PollOTP(options.Otppath)

	if otp != "" {
		log.Println("otp=" + string(otp))

		err = page.Locator("#factor").Fill(otp)
		if err != nil {
			panic(fmt.Sprintf("could not set otp: %v", err))
		}

		err = page.Locator("#VerifyMFA").Click()
		if err != nil {
			panic(fmt.Sprintf("could not click otp: %v", err))
		}
	} else {
		panic("could not get one time password")
	}

	// get balance
	//

	// <a data-qa-button="Details" data-dd-link="Details" data-dd-loc="roundel" data-dd-group="myavivaHomePage" href="/MyPortfolio/ViewDetail?id=A3Acnhvs2bv17h0NKjx1t0s0fhGjFYRBO_3hxv9uIG41&amp;productCode=50010" class="button yellow dd-data-link">Details</a>
	err = page.Locator("[data-qa-button=Details]").Click()
	if err != nil {
		panic(fmt.Sprintf("failed to click on details: %v", err))
	}

	// <p class="a-heading a-heading--0 font-yellow u-margin--top-none" data-qa-field="yourPensionValue">£123,456.72</p>
	balance, err := page.Locator("[data-qa-field=yourPensionValue]").TextContent()
	if err != nil {
		panic(fmt.Sprintf("failed to get balance: %v", err))
	}
	log.Println("balance=" + balance)
	fmt.Println(strings.NewReplacer("£", "", ",", "").Replace(balance))

	bufio.NewWriter(os.Stdout).Flush()
}

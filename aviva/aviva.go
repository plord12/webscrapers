/**

Get aviva balance

*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/plord12/webscrapers/utils"

	"github.com/playwright-community/playwright-go"
)

func main() {

	// defaults from environment
	//
	defaultHeadless := true
	defaultUsername := ""
	defaultPassword := ""
	defaultOtpPath := "otp/aviva"
	defaultOtpCleanCommand := ""
	defaultOtpCommand := ""

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envUsername := os.Getenv("AVIVA_USERNAME"); envUsername != "" {
		defaultUsername = envUsername
	}
	if envPassword := os.Getenv("AVIVA_PASSWORD"); envPassword != "" {
		defaultPassword = envPassword
	}
	if envOtpPath := os.Getenv("OTP_PATH"); envOtpPath != "" {
		defaultOtpPath = envOtpPath
	}
	if envOtpCleanCommand := os.Getenv("OTP_CLEANCOMMAND"); envOtpCleanCommand != "" {
		defaultOtpCleanCommand = envOtpCleanCommand
	}
	if envOtpCommand := os.Getenv("OTP_COMMAND"); envOtpCommand != "" {
		defaultOtpCommand = envOtpCommand
	}

	// arguments
	//
	headless := flag.Bool("headless", defaultHeadless, "Headless mode")
	otpCleanCommand := flag.String("otpcleancommand", defaultOtpCleanCommand, "Command to clean previous one time password")
	otpCommand := flag.String("otpcommand", defaultOtpCommand, "Command to get one time password")
	otpPath := flag.String("otppath", defaultOtpPath, "Path to file containing one time password message")

	username := flag.String("username", defaultUsername, "Aviva username")
	password := flag.String("password", defaultPassword, "Aviva password")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Retrive Aviva balance via web scraping")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $OTP_COMMAND - Command to get one time password")
		fmt.Println("  $OTP_PATH - Path to file containing one time password message")
		fmt.Println("  $AVIVA_USERNAME - Aviva username")
		fmt.Println("  $AVIVA_PASSWORD - Aviva password")
	}

	// parse flags
	//
	flag.Parse()

	// FIX THIS - validate

	// clean from any previous run
	//
	utils.CleanOTP(otpCleanCommand, otpPath)

	// setup
	//
	page := utils.StartCamoufox(headless)
	defer utils.Finish(page)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err := page.Goto("https://www.direct.aviva.co.uk/MyAccount/login", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	// dismiss pop-up
	//
	// <button id="onetrust-accept-btn-handler">Accept all cookies</button>
	page.GetByText("Accept all cookies", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

	log.Printf("Logging in\n")
	// <input aria-required="True" autocomplete="off" class="a-textbox" data-qa-textbox="username" data-val="true" data-val-required="Please enter your username" id="username" maxlength="50" name="username" type="text" value="">
	err = page.Locator("#username").Fill(*username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	// <input aria-required="True" autocomplete="off" class="a-textbox" data-qa-textbox="password" data-val="true" data-val-required="Please enter your password" id="password" maxlength="300" name="password" type="password">
	err = page.Locator("#password").Fill(*password)
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
	utils.FetchOTP(otpCommand)

	// check/poll if otp/aviva exists ... could be via the above command or pushed here elsewhere
	//
	otp := utils.PollOTP(otpPath)

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

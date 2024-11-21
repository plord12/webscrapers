/**

Get nutmeg balance

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

	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

func main() {

	// defaults from environment
	//
	defaultHeadless := true
	defaultUsername := ""
	defaultPassword := ""
	defaultOtpPath := "otp/nutmeg"
	defaultOtpCleanCommand := ""
	defaultOtpCommand := ""

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envUsername := os.Getenv("NUTMEG_USERNAME"); envUsername != "" {
		defaultUsername = envUsername
	}
	if envPassword := os.Getenv("NUTMEG_PASSWORD"); envPassword != "" {
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
	otpCommand := flag.String("otpcommand", defaultOtpCommand, "Command to get one time password")
	otpCleanCommand := flag.String("otpcleancommand", defaultOtpCleanCommand, "Command to clean previous one time password")
	otpPath := flag.String("otppath", defaultOtpPath, "Path to file containing one time password message")

	username := flag.String("username", defaultUsername, "Nutmeg username")
	password := flag.String("password", defaultPassword, "Nutmeg password")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Retrive Nutmeg balance via web scraping")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $OTP_CLEANCOMMAND - Command to clean previous one time password")
		fmt.Println("  $OTP_COMMAND - Command to get one time password")
		fmt.Println("  $OTP_PATH - Path to file containing one time password message")
		fmt.Println("  $NUTMEG_USERNAME - Nutmeg username")
		fmt.Println("  $NUTMEG_PASSWORD - Nutmeg password")
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
	page := utils.StartChromium(headless)
	defer utils.Finish(page)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err := page.Goto("https://authentication.nutmeg.com/login", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	// accept cookies
	//
	page.GetByText("Accept all cookies", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()

	log.Printf("Logging in\n")
	// <input class="input c4ea79246 c882875d6" inputmode="email" name="username" id="username" type="text" aria-label="Email address" value="" required="" autocomplete="off" autocapitalize="none" spellcheck="false" autofocus="">
	err = page.Locator("#username").Fill(*username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	// <input class="input c4ea79246 c2946f7ad" name="password" id="password" type="password" aria-label="Password" required="" autocomplete="current-password" autocapitalize="none" spellcheck="false">
	err = page.Locator("#password").Fill(*password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}
	// <button type="submit" name="action" value="default" class="c0a486a03 c3a925026 cc4e2760d cf0fbb154 c4b20090f" data-action-button-primary="true">Sign in</button>
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign in"}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// attempt to fetch one time password if needed
	//
	utils.FetchOTP(otpCommand)

	// check/poll if otp/aviva exists ... could be via the above command or pushed here elsewhere
	//
	otp := utils.PollOTP(otpPath)

	if otp != "" {
		log.Println("otp=" + string(otp))

		// <label aria-hidden="true" class="cd7843ea8 c6c423b62 c6c2d595a" for="code">Enter the 6-digit code*</label>
		err = page.GetByText("Enter the 6-digit code*", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Fill(otp)
		if err != nil {
			panic(fmt.Sprintf("could not set otp: %v", err))
		}

		// <button type="submit" name="action" value="default" class="c0a486a03 c3a925026 cc4e2760d cf0fbb154 c3a009796" data-action-button-primary="true">Continue</button>
		err = page.GetByText("Continue", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
		if err != nil {
			panic(fmt.Sprintf("could not click otp: %v", err))
		}
	} else {
		panic(fmt.Sprintf("could not get one time password message: %v", err))
	}

	// get balance
	//
	// <span class="_nk-text_1a26s_12 nk-text _nk-amount_16djj_1 _nk-amount--theme-dark_16djj_49 _nk-amount--positive_16djj_10 _nk-amount--no-line-height_16djj_25 _nk-text--style-text-1_1a26s_60 _nk-text--fw-medium_1a26s_133 _nk-text--color-dark_1a26s_160 _nk-text--theme-dark--color-default_1a26s_356 _nk-text--theme-dark--color-dark_1a26s_362 _nk-text--tag-span_1a26s_103 _nk-text--size-xxl_1a26s_106" aria-label="£417,405" role="text" data-qa="portfolio-summary-overview__portfolio-value"><span class="_nk-text_1a26s_12 nk-text _nk-amount__prefix_16djj_46 _nk-text--no-line-height_1a26s_89 _nk-text--no-color_1a26s_172 _nk-text--style-text-1_1a26s_60 _nk-text--fw-medium_1a26s_133 _nk-text--tag-span_1a26s_103 _nk-text--size-xl_1a26s_109">£</span><span class="_nk-text_1a26s_12 nk-text _nk-text--no-line-height_1a26s_89 _nk-text--no-color_1a26s_172 _nk-text--style-text-1_1a26s_60 _nk-text--fw-medium_1a26s_133 _nk-text--tag-span_1a26s_103 _nk-text--size-xxl_1a26s_106" style="font-family: &quot;ivypresto-headline&quot;, sans-serif;">417,405</span></span>
	balance, err := page.Locator("[data-qa=portfolio-summary-overview__portfolio-value]").TextContent()
	if err != nil {
		panic(fmt.Sprintf("failed to get balance: %v", err))
	}
	log.Println("balance=" + balance)
	fmt.Println(strings.NewReplacer("£", "", ",", "").Replace(balance))

	bufio.NewWriter(os.Stdout).Flush()
}

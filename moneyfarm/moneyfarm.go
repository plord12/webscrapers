/**

Get moneyfarm balance

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
	Headless        bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username        string `short:"u" long:"username" description:"Moneyfarm username" env:"MONEYFARM_USERNAME" required:"true"`
	Password        string `short:"p" long:"password" description:"Moneyfarm password" env:"MONEYFARM_PASSWORD" required:"true"`
	Otppath         string `short:"o" long:"otppath" description:"Path to file containing one time password message" default:"otp/moneyfarm" env:"OTP_PATH"`
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
	utils.CleanOTP(options.Otpcleancommand, options.Otppath)

	// setup
	//
	page := utils.StartChromium(options.Headless)
	defer utils.Finish(page)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err = page.Goto("https://app.moneyfarm.com/gb/sign-in", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	// accept cookies
	//
	page.GetByText("Manage preferences", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
	page.GetByText("Save settings", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()

	log.Printf("Logging in\n")
	// <input type="email" id="email" name="email" autocomplete="email" class="sc-dWddBi dbJxuP" value="">
	err = page.Locator("#email").Fill(options.Username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	// <input type="password" id="password" name="password" autocomplete="current-password" class="sc-dWddBi dbJxuP" value="">
	err = page.Locator("#password").Fill(options.Password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}
	// <button data-role="primary" type="submit" data-overlay="false" class="sc-hKgJUU jhVfGS"><span>Sign in</span></button>
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign in"}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// attempt to fetch one time password if needed
	//
	utils.FetchOTP(options.Otpcommand)

	// check/poll if otp exists ... could be via the above command or pushed here elsewhere
	//
	otp := utils.PollOTP(options.Otppath)

	if otp != "" {
		log.Println("otp=" + string(otp))

		// <input class="input c4ea79246 c954c3815 ce0672f58 c3f27bf21 c1a0fa5af" name="code" id="code" type="text" aria-invalid="true" aria-describedby="error-element-code" value="" required="" autocomplete="off" autocapitalize="none" spellcheck="false" autofocus=""><div class="cd7843ea8 js-required c6c423b62 c6c2d595a" data-dynamic-label-for="code" aria-hidden="true">Enter the 6-digit code*</div></div>
		err = page.Locator("#code").Fill(otp)
		if err != nil {
			panic(fmt.Sprintf("could not set otp: %v", err))
		}

		// <button type="submit" name="action" value="default" class="c0a486a03 c3a925026 cc4e2760d cf0fbb154 c3a009796" data-action-button-primary="true">Continue</button>
		err = page.GetByText("Continue", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
		if err != nil {
			panic(fmt.Sprintf("could not click otp: %v", err))
		}
	} else {
		panic("could not get one time password")
	}

	// get balance
	//
	balance, err := page.GetByText("£").First().TextContent()
	if err != nil {
		panic(fmt.Sprintf("failed to get balance: %v", err))
	}
	log.Println("balance=" + balance)
	fmt.Println(strings.NewReplacer("£", "", ",", "").Replace(balance))

	bufio.NewWriter(os.Stdout).Flush()
}

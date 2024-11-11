/**

Get moneyfarm balance

*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	stealth "github.com/jonfriesen/playwright-go-stealth"
	"github.com/playwright-community/playwright-go"
)

func main() {

	// defaults from environment
	//
	defaultHeadless := true
	defaultUsername := ""
	defaultPassword := ""
	defaultOtpPath := "otp/moneyfarm"
	defaultOtpCleanCommand := ""
	defaultOtpCommand := ""

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envUsername := os.Getenv("MONEYFARM_USERNAME"); envUsername != "" {
		defaultUsername = envUsername
	}
	if envPassword := os.Getenv("MONEYFARM_PASSWORD"); envPassword != "" {
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

	username := flag.String("username", defaultUsername, "Moneyfarm username")
	password := flag.String("password", defaultPassword, "Moneyfarm password")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Retrive Moneyfarm balance via web scraping")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $OTP_CLEANCOMMAND - Command to clean previous one time password")
		fmt.Println("  $OTP_COMMAND - Command to get one time password")
		fmt.Println("  $OTP_PATH - Path to file containing one time password message")
		fmt.Println("  $MONEYFARM_USERNAME - Moneyfarm username")
		fmt.Println("  $MONEYFARM_PASSWORD - Moneyfarm password")
	}

	// parse flags
	//
	flag.Parse()

	// FIX THIS - validate

	// clean from any previous run
	//
	if *otpCleanCommand != "" {
		log.Printf("Running %s to clean old one time password\n", *otpCleanCommand)
		command := strings.Split(*otpCleanCommand, " ")
		exec.Command(command[0], command[1:]...).Run()
	}
	os.Remove(*otpPath)

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
	log.Printf("Starting login\n")
	_, err = page.Goto("https://app.moneyfarm.com/gb/sign-in", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not goto url: %v", err)
	}

	// accept cookies
	//
	page.GetByText("OK, I agree", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()

	log.Printf("Logging in\n")
	// <input type="email" id="email" name="email" autocomplete="email" class="sc-dWddBi dbJxuP" value="">
	err = page.Locator("#email").Fill(*username)
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not get username: %v", err)
	}
	// <input type="password" id="password" name="password" autocomplete="current-password" class="sc-dWddBi dbJxuP" value="">
	err = page.Locator("#password").Fill(*password)
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not get password: %v", err)
	}
	// <button data-role="primary" type="submit" data-overlay="false" class="sc-hKgJUU jhVfGS"><span>Sign in</span></button>
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign in"}).Click()
	if err != nil {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not click: %v", err)
	}

	// attempt to fetch one time password if needed
	//
	if *otpCommand != "" {
		log.Printf("Running %s to get one time password\n", *otpCommand)
		for i := 0; i < 30; i++ {
			command := strings.Split(*otpCommand, " ")
			cmd := exec.Command(command[0], command[1:]...)
			err := cmd.Run()
			if err != nil {
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
	}

	// check/poll if otp exists ... could be via the above command or pushed here elsewhere
	//
	otp := ""
	for i := 0; i < 30; i++ {
		_, err := os.Stat(*otpPath)
		if errors.Is(err, os.ErrNotExist) {
			time.Sleep(2 * time.Second)
		} else {
			// read otp
			//
			data, err := os.ReadFile(*otpPath)
			if err == nil {
				r := regexp.MustCompile(".*([0-9][0-9][0-9][0-9][0-9][0-9]).*")
				match := r.FindStringSubmatch(string(data))
				if len(match) != 2 {
					browser.Close()
					pw.Stop()
					log.Fatalf("could not parse one time password message: %v", err)
				} else {
					otp = match[1]
				}
			}
			os.Remove(*otpPath)
			break
		}
	}

	os.Remove(*otpPath)

	if otp != "" {
		log.Println("otp=" + string(otp))

		// <input class="input c4ea79246 c954c3815 ce0672f58 c3f27bf21 c1a0fa5af" name="code" id="code" type="text" aria-invalid="true" aria-describedby="error-element-code" value="" required="" autocomplete="off" autocapitalize="none" spellcheck="false" autofocus=""><div class="cd7843ea8 js-required c6c423b62 c6c2d595a" data-dynamic-label-for="code" aria-hidden="true">Enter the 6-digit code*</div></div>
		err = page.Locator("#code").Fill(otp)
		if err != nil {
			browser.Close()
			pw.Stop()
			log.Fatalf("could not set otp: %v", err)
		}

		// <button type="submit" name="action" value="default" class="c0a486a03 c3a925026 cc4e2760d cf0fbb154 c3a009796" data-action-button-primary="true">Continue</button>
		err = page.GetByText("Continue", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
		if err != nil {
			browser.Close()
			pw.Stop()
			log.Fatalf("could not click otp: %v", err)
		}
	} else {
		browser.Close()
		pw.Stop()
		log.Fatalf("could not get one time password message: %v", err)
	}

	// get balance
	//
	// <span aria-hidden="false" class="sc-jcRCNh ieovWt">£92,276.76</span>
	balance, err := page.Locator("[class=\"sc-jcRCNh ieovWt\"]").First().TextContent()
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

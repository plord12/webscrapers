/**

Load next batch of kindle unlimited books

*/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless        bool     `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username        string   `short:"u" long:"username" description:"Amazon username" env:"AMAZON_USERNAME" required:"true"`
	Password        string   `short:"p" long:"password" description:"Amazon password" env:"AMAZON_PASSWORD" required:"true"`
	Otppath         string   `short:"o" long:"otppath" description:"Path to file containing one time password message" default:"otp/jpmorganpi" env:"OTP_PATH"`
	Otpcommand      string   `short:"c" long:"otpcommand" description:"Command to get one time password" env:"OTP_COMMAND"`
	Otpcleancommand string   `short:"l" long:"otpcleancommand" description:"Command to clean previous one time password" env:"OTP_CLEANCOMMAND"`
	Return          bool     `short:"d" long:"return" description:"Return existing borrowed books" env:"RETURN"`
	Maximum         int      `short:"m" long:"maximum" description:"Maxiumum number of books to borrow" default:"1" env:"MAXIMUMN"`
	Section         []string `short:"s" long:"section" description:"Kindle unlimited sections" default:"Continue series you’ve started with Kindle Unlimited" env:"SECTION"`
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

	// get existing books
	//
	command := strings.Split("calibredb list --for-machine --fields=all", " ")
	out, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		panic(fmt.Sprintf("could not exec calibredb: %v", err))
	}
	type Book struct {
		Title       string
		Identifiers map[string]string
	}
	var existingbooks []Book
	err = json.Unmarshal([]byte(out), &existingbooks)
	if err != nil {
		panic(fmt.Sprintf("could not unmarshal json: %v", err))
	}

	// setup
	//
	page := utils.StartCamoufox(options.Headless)
	defer utils.Finish(page)
	page.SetViewportSize(1920, 1080)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err = page.Goto("https://www.amazon.co.uk", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}
	page.GetByText("Decline", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

	_, err = page.Goto("https://www.amazon.co.uk/hz/mycd/digital-console/contentlist/kuAll/dateDsc/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	time.Sleep(3 * time.Second)

	log.Printf("Logging in\n")
	err = page.Locator("#ap_email").Fill(options.Username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	err = page.Locator("#ap_password").Fill(options.Password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}

	err = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign in"}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// attempt to fetch one time password if needed
	//
	/*
		utils.FetchOTP(options.Otpcommand)
		otp := utils.PollOTP(options.Otppath)

		if otp != "" {
			log.Println("otp=" + string(otp))

			err = page.Locator("#code-label").Fill(otp)
			if err != nil {
				panic(fmt.Sprintf("could not set otp: %v", err))
			}

			err = page.GetByText("Continue", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
			if err != nil {
				panic(fmt.Sprintf("could not click otp: %v", err))
			}
		} else {
			panic(fmt.Sprintf("could not get one time password message: %v", err))
		}
	*/

	time.Sleep(5 * time.Second) // FIX ... poll ?

	if options.Return {
		for {
			// could make use of [id^="RETURN_CONTENT_ACTION_"] & visible
			borrowed, err := page.Locator(".action_button", playwright.PageLocatorOptions{HasText: "Return this book"}).Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).All()
			if err == nil && len(borrowed) > 0 {
				book := page.Locator(".action_button", playwright.PageLocatorOptions{HasText: "Return this book"}).Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).First()
				log.Printf("Returning item\n")
				book.ScrollIntoViewIfNeeded()
				time.Sleep(1 * time.Second)
				book.Click()
				time.Sleep(1 * time.Second)
				page.GetByRole("button").Filter(playwright.LocatorFilterOptions{HasText: "Return this book"}).First().Click()
				time.Sleep(1 * time.Second)
				page.Locator("#notification-close").First().Click()
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
	}

	log.Printf("Starting new books\n")
	for _, section := range options.Section {
		_, err = page.Goto("https://www.amazon.co.uk/kindle-dbs/hz/subscribe/ku?ref=ebooks_dsk_uk_kindle_unlimited", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			panic(fmt.Sprintf("could not goto url: %v", err))
		}
		page.Locator("h2").Filter(playwright.LocatorFilterOptions{HasText: section}).First().ScrollIntoViewIfNeeded()
		page.Locator("[aria-label=\"See more " + section + "\"]").First().Click()
		time.Sleep(1 * time.Second)

		booksborrowed := 0

		base := page.URL()

		newbooks, err := page.Locator(".s-no-outline").Or(page.Locator(".browse-grid-view-link")).All()
		//fmt.Fprintf(os.Stderr, "err=%v len=%d\n", err, len(newbooks))
		if err == nil && len(newbooks) > 0 {
			for _, newbook := range newbooks {
				href, _ := newbook.GetAttribute("href")
				//fmt.Fprintf(os.Stderr, "href=%s\n", href)
				// /gp/product/B003YUCEB6?ref_=dbs_b_r_brws_recs_l_p1_0&storeType=ebooks
				// /Law-Maker-Aristocrats-London-Book-ebook/dp/B0FJCDSJDT/r...
				var idRegex, _ = regexp.Compile(".*/(B0[0-9A-Z]{8}).*")
				id := idRegex.FindStringSubmatch(href)
				if len(id) > 1 {

					borrow := true
					for _, book := range existingbooks {
						if id[1] == book.Identifiers["mobi-asin"] {
							log.Printf("id %s already downloaded\n", book.Identifiers["mobi-asin"])
							borrow = false
							break
						}
					}

					if borrow {
						_, err = page.Goto("https://www.amazon.co.uk"+href, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
						if err != nil {
							panic(fmt.Sprintf("could not goto url: %v", err))
						}
						time.Sleep(1 * time.Second)
						page.Locator("#borrow-button-announce").First().Click()
						log.Printf("Borrowed %s\n", id[1])

						time.Sleep(1 * time.Second)
						_, err = page.Goto(base, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
						if err != nil {
							panic(fmt.Sprintf("could not goto url: %v", err))
						}

						booksborrowed++
					}
				}

				// can't borrow more than 20
				//
				if booksborrowed >= (options.Maximum/len(options.Section)) || booksborrowed >= 20 {
					break
				}
			}
		}
	}
	// FIX THIS - can hit next

	// FIX THIS - also try "Our favorites"

	time.Sleep(5 * time.Second)

	// now download to device
	//
	// ideally, only download books we've just borrowed
	//
	_, err = page.Goto("https://www.amazon.co.uk/hz/mycd/digital-console/contentlist/kuAll/dateDsc/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}
	time.Sleep(1 * time.Second)
	page.Locator("#SELECT-ALL").First().Click()
	time.Sleep(1 * time.Second)
	page.Locator(".action_button").Filter(playwright.LocatorFilterOptions{HasText: "Deliver to device"}).First().Click()
	time.Sleep(1 * time.Second)
	page.GetByRole("checkbox").First().Check()

	page.GetByText("Make Changes").First().Click()

	time.Sleep(3 * time.Second)

	bufio.NewWriter(os.Stdout).Flush()
}

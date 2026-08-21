/**

Load next batch of kindle unlimited books

*/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
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
	Headless        bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username        string `short:"u" long:"username" description:"Amazon username" env:"AMAZON_USERNAME" required:"true"`
	Password        string `short:"p" long:"password" description:"Amazon password" env:"AMAZON_PASSWORD" required:"true"`
	Otppath         string `short:"o" long:"otppath" description:"Path to file containing one time password message" default:"otp/amazon" env:"OTP_PATH"`
	Otpcommand      string `short:"c" long:"otpcommand" description:"Command to get one time password" env:"OTP_COMMAND"`
	Otpcleancommand string `short:"l" long:"otpcleancommand" description:"Command to clean previous one time password" env:"OTP_CLEANCOMMAND"`
	Return          bool   `short:"d" long:"return" description:"Return existing borrowed books" env:"RETURN"`
	Maximum         int    `short:"m" long:"maximum" description:"Maxiumum number of books to borrow" default:"1" env:"MAXIMUMN"`
	Test            bool   `short:"t" long:"test" description:"Test mode" env:"TESTMODE"`
	Search          string `short:"s" long:"search" description:"Kindle search" env:"SEARCH"`
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

	// list of books borrowed in this session
	//
	borrowed := make(map[string]bool)

	// setup
	//
	page := utils.StartCamoufox(options.Headless)
	newContext, err := page.Context().Browser().NewContext()
	if err != nil {
		panic(fmt.Sprintf("could not open new page: %v", err))
	}
	page1, err := newContext.NewPage()
	if err != nil {
		panic(fmt.Sprintf("could not open new page: %v", err))
	}
	defer utils.Finish(page1)
	page1.SetViewportSize(1920, 1080)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err = page1.Goto("https://www.amazon.co.uk", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}
	page1.GetByText("Decline", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(2000.0)})

	_, err = page1.Goto("https://www.amazon.co.uk/hz/mycd/digital-console/contentlist/kuAll/dateDsc/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	time.Sleep(3 * time.Second)

	log.Printf("Logging in\n")

	err = page1.Locator("#ap_email_login").Fill(options.Username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	err = page1.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Continue"}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	/*
		err = page1.Locator("#ap_email").Fill(options.Username)
		if err != nil {
			panic(fmt.Sprintf("could not get username: %v", err))
		}
	*/

	err = page1.Locator("#ap_password").Fill(options.Password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}

	err = page1.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign in", Exact: playwright.Bool(true)}).Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// check for app approval
	//
	err = page1.GetByText("For your security, approve the notification sent to:").WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(5000.0)})
	if err == nil {
		log.Println("Need app approval")
		time.Sleep(5 * 60 * time.Second)
	}

	// check for one time password if needed
	//
	err = page1.Locator("#input-box-otp").WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(5000.0)})
	if err == nil {
		log.Println("Need OTP")
		utils.FetchOTP(options.Otpcommand)
		otp := utils.PollOTP(options.Otppath)

		if otp != "" {
			log.Println("otp=" + string(otp))

			err = page1.Locator("#input-box-otp").Fill(otp)
			if err != nil {
				panic(fmt.Sprintf("could not set otp: %v", err))
			}

			err = page1.GetByText("Submit code", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
			if err != nil {
				panic(fmt.Sprintf("could not click otp: %v", err))
			}
		} else {
			panic(fmt.Sprintf("could not get one time password message: %v", err))
		}

	}
	time.Sleep(2 * time.Second) // FIX ... poll ?

	if options.Return {
		log.Printf("Starting returning\n")
		for {
			// could make use of [id^="RETURN_CONTENT_ACTION_"] & visible
			borrowed, err := page1.Locator(".action_button", playwright.PageLocatorOptions{HasText: "Return this book"}).Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).All()
			if err == nil && len(borrowed) > 0 {
				book := page1.Locator(".action_button", playwright.PageLocatorOptions{HasText: "Return this book"}).Filter(playwright.LocatorFilterOptions{Visible: playwright.Bool(true)}).First()
				log.Printf("Returning item\n")
				book.ScrollIntoViewIfNeeded()
				time.Sleep(1 * time.Second)
				book.Click()
				time.Sleep(1 * time.Second)
				if options.Test {
					log.Printf("TEST: Skipping return this book\n")
					page1.GetByRole("button").Filter(playwright.LocatorFilterOptions{HasText: "Cancel"}).First().Click()
					break
				} else {
					page1.GetByRole("button").Filter(playwright.LocatorFilterOptions{HasText: "Return this book"}).First().Click()
				}
				time.Sleep(1 * time.Second)
				page1.Locator("#notification-close").First().Click()
				time.Sleep(1 * time.Second)
			} else {
				log.Printf("Finished returning\n")
				break
			}
		}
	}

	// FIX THIS - add search eg https://www.amazon.co.uk/s?k=ted+bun&i=kindle-unlimited
	//
	// look for action.asin not downloaded and "Read now" button
	//
	// <form method="post" action="/api/bifrost/acquisitions/v1/asins/B0GPFKGKY2?x-client-id=search-dbs&amp;ref_=sr_rn_be_tp">
	// <input type="hidden" name="items[0].action.actionType" value="Borrow">
	// <input type="hidden" name="items[0].action.program.name" value="KU">
	// <input type="hidden" name="items[0].action.program.programCode" value="KINDLE_UNLIMITED">
	// <input type="hidden" name="csrf" value="g0rDipDCoT8YDlZBHQTjEAZPwkSYoMKZQiomr/5bIsmvAAAAAQAAAABqc0+IcmF3AAAAAKs+FBXVfD4nuL9rqj+OIQ==">
	// <input type="hidden" name="items[0].searchContext.queryId" value="1785941896">
	// <input type="hidden" name="items[0].action.asin" value="B0GPFKGKY2">
	// <input type="hidden" name="items[0].searchContext.searchRank" value="1-2">
	// <input type="hidden" name="items[0].action.program.channelCode" value="ALL_YOU_CAN_READ">
	// <span class="a-button a-button-base" id="a-autoid-2">
	// <span class="a-button-inner">
	// <input aria-label="Read now" class="a-button-input" type="submit">
	// <span class="a-button-text" aria-hidden="true" id="a-autoid-2-announce">Read now</span>
	// </span>
	// </span>
	// </form>
	//
	// Hit Next if possible
	//
	// <a href="/s?k=ted+bun&amp;i=kindle-unlimited&amp;page=2&amp;xpid=MEneInFYNmD61&amp;qid=1785941896&amp;ref=sr_pg_1" role="button" tabindex="0" aria-label="Go to next page, page 2" class="s-pagination-item s-pagination-next s-pagination-button s-pagination-button-accessibility s-pagination-separator">Next<svg xmlns="http://www.w3.org/2000/svg" width="8" height="12" viewBox="0 0 8 12" focusable="false" aria-hidden="true"><path d="M2.126.35a1.28 1.28 0 00-1.761 0 1.165 1.165 0 000 1.695L4.478 6 .365 9.955a1.165 1.165 0 000 1.694 1.28 1.28 0 001.76 0L8 6 2.126.35z"></path></svg></a>
	//

	if len(options.Search) > 0 {

		_, err = page1.Goto("https://www.amazon.co.uk/s?k="+html.EscapeString(options.Search)+"&i=kindle-unlimited", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			panic(fmt.Sprintf("could not goto url: %v", err))
		}
		time.Sleep(3 * time.Second)

		booksborrowed := 0

		for screen := 0; screen < 5; screen++ {
			newbooks, err := page1.Locator("form", playwright.PageLocatorOptions{HasText: "Read now"}).All()
			//fmt.Fprintf(os.Stderr, "err=%v len=%d\n", err, len(newbooks))
			if err == nil && len(newbooks) > 0 {
				for _, newbook := range newbooks {
					// fmt.Fprintf(os.Stderr, "%v\n", newbook)
					b, err := newbook.Locator("[name=\"items[0].action.asin\"]").GetAttribute("value")
					if err != nil {
						break
					}
					borrow := true
					for _, book := range existingbooks {
						if b == book.Identifiers["mobi-asin"] {
							log.Printf("id %s already downloaded\n", b)
							borrow = false
							break
						}
					}

					if borrowed[b] {
						log.Printf("id %s already borrowed in this session\n", b)
						borrow = false
					}

					if borrow {
						// borrow in a new window
						//
						// <a aria-hidden="true" class="a-link-normal s-no-outline" tabindex="-1" href="/Bare-Against-Black-Dog-Ted-ebook/dp/B0GPFKGKY2/ref=sr_1_1?dib=eyJ2IjoiMSJ9.nJ5YCnN6gFkxXqahuDp7dT8wqYGlQhim-kgOJ6arZlAlMGH_RHscIJ_91KV4sAjtRYbdwIEusAlwtwLV1OqneNR6pkQ6T4P4PbftGzCNcI-1-Pyup5xaB9jDnKJ72S8p4PfAs4RQak7xnxvqODoTKuFAUrWCUNlQcqmCJ-2KYXcn4wTW8VR6dre1v8UPIqezX0QhfCVwmGgzLjRNpdWMlgmKGHLlYhZRO5IIOpuPzC8.vpv-iOrPWmXQQ0npJZHMZ8NEAd4kPgEllsDHqp-14aY&amp;dib_tag=se&amp;keywords=ted+bun&amp;qid=1785949905&amp;s=digital-text&amp;sr=1-1"><div class="a-section aok-relative s-image-fixed-height"><img class="s-image" src="https://m.media-amazon.com/images/I/71TEt11+NJL._AC_UY218_.jpg" srcset="https://m.media-amazon.com/images/I/71TEt11+NJL._AC_UY218_.jpg 1x, https://m.media-amazon.com/images/I/71TEt11+NJL._AC_UY327_FMwebp_QL65_.jpg 1.5x, https://m.media-amazon.com/images/I/71TEt11+NJL._AC_UY436_FMwebp_QL65_.jpg 2x, https://m.media-amazon.com/images/I/71TEt11+NJL._AC_UY545_FMwebp_QL65_.jpg 2.5x, https://m.media-amazon.com/images/I/71TEt11+NJL._AC_UY654_FMwebp_QL65_.jpg 3x" alt="Bare Against the Black Dog" aria-hidden="true" data-image-index="1" data-image-load="" data-image-latency="s-product-image" data-image-source-density="1"></div></a>
						href, err := page1.Locator("[href*=\"/" + b + "/\"]").First().GetAttribute("href")
						if err == nil {
							page2, err := newContext.NewPage()
							if err != nil {
								panic(fmt.Sprintf("could not open new page: %v", err))
							}
							defer utils.Finish(page2)
							page2.SetViewportSize(1920, 1080)

							_, err = page2.Goto("https://www.amazon.co.uk"+href, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
							if err != nil {
								log.Printf("could not goto url: %v", err)
								break
							}
							time.Sleep(1 * time.Second)
							if options.Test {
								log.Printf("TEST: Skipping borrow book\n")
							} else {
								page2.Locator("#borrow-button-announce").First().Click()
							}
							booksborrowed++
							log.Printf("Borrowed %s (%d)\n", b, booksborrowed)
							borrowed[b] = true

							err = page2.Close()
							if err != nil {
								log.Printf("could not close page: %v", err)
							}
						}
						if booksborrowed >= options.Maximum {
							break
						}
					}
				}
			}
			if booksborrowed >= options.Maximum {
				break
			}
			// try next page
			page1.Evaluate(`window.scrollBy(0, 30000)`)
			err = page1.GetByText("Next", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Or(page1.GetByText("See all results", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)})).Click()
			time.Sleep(1 * time.Second)
			if err != nil {
				break
			}

		}

	} else {

		_, err = page1.Goto("https://www.amazon.co.uk/kindle-dbs/hz/bookshelf?ref_=sv_kindlebooks_store_1", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			panic(fmt.Sprintf("could not goto url: %v", err))
		}
		time.Sleep(3 * time.Second)

		sections, err := page1.Locator(".card-header-cta-text").All()
		booksborrowed := 0

		for _, section := range sections {
			log.Printf("Next section %v\n", section)
			err = section.Click()
			if err != nil {
				panic(fmt.Sprintf("could not goto section: %v", err))
			}
			time.Sleep(3 * time.Second)
			/*
				// scroll into view until visible
				//
				for i := 0; i < 10; i++ {
					v, _ := page.Locator("[aria-label=\"See more " + section + "\"]").First().IsVisible()
					if v {
						break
					}
					page.Evaluate(`window.scrollBy(0, 400)`)
				}

				page.Locator("[aria-label=\"See more " + section + "\"]").First().Click()
				time.Sleep(1 * time.Second)
			*/

			for screen := 0; screen < 3; screen++ {
				newbooks, err := page1.Locator(".s-no-outline").Or(page1.Locator(".browse-grid-view-link")).All()
				//fmt.Fprintf(os.Stderr, "err=%v len=%d\n", err, len(newbooks))
				if err == nil && len(newbooks) > 0 {
					for _, newbook := range newbooks {
						href, _ := newbook.GetAttribute("href")
						//fmt.Fprintf(os.Stderr, "href=%s\n", href)
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
							if borrowed[id[1]] {
								log.Printf("id %s already borrowed in this session\n", id[1])
								borrow = false
							}

							if borrow {
								// borrow in a new window
								page2, err := newContext.NewPage()
								if err != nil {
									panic(fmt.Sprintf("could not open new page: %v", err))
								}
								defer utils.Finish(page2)
								page2.SetViewportSize(1920, 1080)

								_, err = page2.Goto("https://www.amazon.co.uk"+href, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
								if err != nil {
									log.Printf("could not goto url: %v", err)
									break
								}
								time.Sleep(1 * time.Second)
								if options.Test {
									log.Printf("TEST: Skipping borrow book\n")
								} else {
									page2.Locator("#borrow-button-announce").First().Click()
								}
								booksborrowed++
								log.Printf("Borrowed %s (%d)\n", id[1], booksborrowed)
								borrowed[id[1]] = true

								err = page2.Close()
								if err != nil {
									log.Printf("could not close page: %v", err)
								}
							}
						}
						if booksborrowed >= options.Maximum {
							break
						}
					}
				}
				if booksborrowed >= options.Maximum {
					break
				}
				// try next page
				page1.Evaluate(`window.scrollBy(0, 30000)`)
				err = page1.GetByText("Next", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Or(page1.GetByText("See all results", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)})).Click()
				time.Sleep(1 * time.Second)
				if err != nil {
					break
				}
			}
			_, err = page1.Goto("https://www.amazon.co.uk/kindle-dbs/hz/bookshelf?ref_=sv_kindlebooks_store_1", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			if err != nil {
				panic(fmt.Sprintf("could not goto url: %v", err))
			}
			if booksborrowed >= options.Maximum {
				break
			}
		}
	}

	time.Sleep(5 * time.Second)

	// now download to device
	//
	// ideally, only download books we've just borrowed
	//
	for download := 0; download < 2; download++ {
		_, err = page1.Goto("https://www.amazon.co.uk/hz/mycd/digital-console/contentlist/kuAll/dateDsc/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		if err != nil {
			panic(fmt.Sprintf("could not goto url: %v", err))
		}
		time.Sleep(1 * time.Second)
		page1.Locator("#SELECT-ALL").First().Click()
		time.Sleep(1 * time.Second)
		page1.Locator(".action_button").Filter(playwright.LocatorFilterOptions{HasText: "Deliver to device"}).First().Click()
		time.Sleep(1 * time.Second)
		page1.GetByRole("checkbox").First().Check()

		if options.Test {
			log.Printf("TEST: Skipping deliver\n")
		} else {
			page1.GetByText("Make Changes").First().Click()
			time.Sleep(1 * time.Second)
			page1.GetByText("Close").First().Click()
			time.Sleep(1 * time.Second)
		}
	}

	bufio.NewWriter(os.Stdout).Flush()
}

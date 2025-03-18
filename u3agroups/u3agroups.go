/**

Do a search for u3a groups and send a message to the group leader

*/

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Search   string `short:"s" long:"search" description:"Search term" env:"U3AGROUPS_SEARCH" required:"true"`
	Name     string `short:"n" long:"name" description:"Message name" env:"U3AGROUPS_NAME" required:"true"`
	Email    string `short:"m" long:"email" description:"Message email" env:"U3AGROUPS_EMAIL" required:"true"`
	Subject  string `short:"u" long:"subject" description:"Message subject" env:"U3AGROUPS_SUBJECT" required:"true"`
	Message  string `short:"g" long:"message" description:"Message" env:"U3AGROUPS_MESSAGE" required:"true"`
	Send     bool   `short:"d" long:"send" description:"Send message" env:"U3AGROUPS_SEND"`
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
	page := utils.StartChromium(options.Headless)
	defer utils.Finish(page)

	// main page & search
	//
	log.Printf("Starting search\n")
	_, err = page.Goto("https://u3asites.org.uk/oversights/groups/groupsearch.php", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}
	err = page.Locator("[name=searchterm]").Fill(options.Search)
	if err != nil {
		panic(fmt.Sprintf("could not get search: %v", err))
	}
	err = page.Locator("[type=submit]").Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// get all links - will be hidden but we don't care
	//
	links, err := page.Locator("a").All()
	if err != nil {
		panic("Could not find links")
	}
	for _, link := range links {
		url, err := link.GetAttribute("href")
		if err != nil {
			panic("Could not find group link")
		}
		if strings.Contains(url, "u3asite.uk") {
			fmt.Println(url)

			page1, err := page.Context().Browser().NewPage()
			if err != nil {
				panic("Could not open new page")
			}
			page1.SetDefaultTimeout(1000.0)
			_, err = page1.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			if err != nil {
				panic("Could not open new page")
			}
			form := page1.GetByTitle("Opens message form", playwright.PageGetByTitleOptions{Exact: playwright.Bool(true)}).First()
			if form != nil {
				form.Click()
			} else {
				fmt.Println("No form")
			}

			page1.Locator("[name=returnName]").Fill(options.Name)
			page1.Locator("[name=returnEmail]").Fill(options.Email)
			page1.Locator("[name=messageSubject]").Fill(options.Subject)
			page1.Locator("[name=messageText]").Fill(options.Message)

			if options.Send {
				err = page1.Locator("[name=sendEmail]").Click()
				if err != nil {
					fmt.Printf("could not click send Email: %f\n", err)
				}
				fmt.Println("Sent")
			}

			time.Sleep(1 * time.Second)

			page1.Close()
		}

	}

	log.Printf("Done\n")

	time.Sleep(5 * time.Second)
}

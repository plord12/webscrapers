/**

Get mycause uk events

*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

type Options struct {
	Headless bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username string `short:"u" long:"username" description:"My cause uk username" env:"MYCAUSEUK_USERNAME" required:"true"`
	Password string `short:"p" long:"password" description:"My cause uk password" env:"MYCAUSEUK_PASSWORD" required:"true"`
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

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err = page.Goto("https://mycauseuk.paamapplication.co.uk/mycauseuk/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	log.Printf("Logging in\n")
	// <input type="text" name="username" id="username">
	err = page.Locator("#username").Fill(options.Username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	// <input type="password" name="password" id="password">
	err = page.Locator("#password").Fill(options.Password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}
	// <input type="submit" name="submit" value="login" class="submit_btn" style="margin-left:73px;">
	err = page.Locator(".submit_btn").Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}

	// events page
	//
	_, err = page.Goto("https://mycauseuk.paamapplication.co.uk/mycauseuk/members/events.php", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	// dump table
	//
	// <table id="eventstable" class="uevents"><tbody><tr>	<th>Event</th>	<th>Date</th>	<th>Location</th>	<th width="100">&nbsp;</th></tr>            <tr>
	//            <td><span class="name">Bearded Theory</span></td>
	//            <td>22nd May  - 27th May</td>
	//            <td> Catton Hall, Walton-on-Trent, Derbyshire, England, UK</td>
	//            <td><span class="status"></span></td>
	//            <td><span class="apply"><span class="full" style="display: block;width: 100%;">Soon</span></span></td>
	//        </tr>
	//       </tbody></table>

	table, err := page.Locator("#eventstable").Locator("tr").All()
	if err != nil {
		panic(fmt.Sprintf("could not get table: %v", err))
	}
	for _, row := range table {
		columns, _ := row.Locator("td").All()
		for _, column := range columns {
			t, _ := column.TextContent()
			fmt.Print(t + "|")
		}
		fmt.Println()
	}

	bufio.NewWriter(os.Stdout).Flush()
}

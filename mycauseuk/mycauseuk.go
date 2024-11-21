/**

Get mycause uk events

*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/playwright-community/playwright-go"
	"github.com/plord12/webscrapers/utils"
)

func main() {

	// defaults from environment
	//
	defaultHeadless := true
	defaultUsername := ""
	defaultPassword := ""

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envUsername := os.Getenv("MYCAUSEUK_USERNAME"); envUsername != "" {
		defaultUsername = envUsername
	}
	if envPassword := os.Getenv("MYCAUSEUK_PASSWORD"); envPassword != "" {
		defaultPassword = envPassword
	}

	// arguments
	//
	headless := flag.Bool("headless", defaultHeadless, "Headless mode")

	username := flag.String("username", defaultUsername, "My cause uk username")
	password := flag.String("password", defaultPassword, "My cause uk password")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Retrive mycause uk events via web scraping")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $MYCAUSEUK_USERNAME - My cause uk username")
		fmt.Println("  $MYCAUSEUK_PASSWORD - My cause uk password")
	}

	// parse flags
	//
	flag.Parse()

	// FIX THIS - validate

	// setup
	//
	page := utils.StartChromium(headless)
	defer utils.Finish(page)

	// main page & login
	//
	log.Printf("Starting login\n")
	_, err := page.Goto("https://mycauseuk.paamapplication.co.uk/mycauseuk/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	log.Printf("Logging in\n")
	// <input type="text" name="username" id="username">
	err = page.Locator("#username").Fill(*username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	// <input type="password" name="password" id="password">
	err = page.Locator("#password").Fill(*password)
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

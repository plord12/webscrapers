/**

sping octopus wheel

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

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envUsername := os.Getenv("OCTOPUS_USERNAME"); envUsername != "" {
		defaultUsername = envUsername
	}
	if envPassword := os.Getenv("OCTOPUS_PASSWORD"); envPassword != "" {
		defaultPassword = envPassword
	}

	// arguments
	//
	headless := flag.Bool("headless", defaultHeadless, "Headless mode")

	username := flag.String("username", defaultUsername, "Octopus username")
	password := flag.String("password", defaultPassword, "Octopus password")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Spin octopus wheel of fortune via web scraping")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $OCTOPUS_USERNAME - Octopus username")
		fmt.Println("  $OCTOPUS_PASSWORD - Octopus password")
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
	_, err := page.Goto("https://octopus.energy/login/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}

	log.Printf("Logging in\n")
	err = page.Locator("#id_username").Fill(*username)
	if err != nil {
		panic(fmt.Sprintf("could not get username: %v", err))
	}
	err = page.Locator("#id_password").Fill(*password)
	if err != nil {
		panic(fmt.Sprintf("could not get password: %v", err))
	}
	err = page.Locator(".button").Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}
	err = page.Locator(".jAWbYk").Click()
	if err != nil {
		panic(fmt.Sprintf("could not click: %v", err))
	}
	split := strings.Split(page.URL(), "/")
	var account string
	if len(split) > 6 {
		account = split[6]
	} else {
		panic("could not get account")
	}

	page.SetDefaultTimeout(5000)

	log.Println("Spinning electricity for " + account)
	// electricity
	//
	_, err = page.Goto("https://octopus.energy/dashboard/new/accounts/"+account+"/wheel-of-fortune/electricity", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}
	page.Locator(".wheel").Click()

	log.Println("Spinning gas for " + account)
	// electricity
	//
	_, err = page.Goto("https://octopus.energy/dashboard/new/accounts/"+account+"/wheel-of-fortune/gas", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		panic(fmt.Sprintf("could not goto url: %v", err))
	}
	page.Locator(".wheel").Click()

	bufio.NewWriter(os.Stdout).Flush()
}

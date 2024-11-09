/**
Home Assistant screenshot tool

Requires :

* Username / password
* URL
* CCS selector

*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/playwright-community/playwright-go"
)

type request struct {
	Url      string `json:"url"`
	Css      string `json:"css"`
	Filename string `json:"filename"`
}

var headless *bool
var username *string
var password *string

func main() {

	// defaults from environment
	//
	defaultHeadless := true
	defaultRestPort := int(0)
	defaultUsername := ""
	defaultPassword := ""
	defaultUrl := ""
	defaultCss := ""
	defaultPath := "output.png"

	if envHeadless := os.Getenv("HEADLESS"); envHeadless != "" {
		defaultHeadless, _ = strconv.ParseBool(envHeadless)
	}
	if envRestPort := os.Getenv("HA_RESTPORT"); envRestPort != "" {
		defaultRestPort, _ = strconv.Atoi(envRestPort)
	}
	if envUsername := os.Getenv("HA_USERNAME"); envUsername != "" {
		defaultUsername = envUsername
	}
	if envPassword := os.Getenv("HA_PASSWORD"); envPassword != "" {
		defaultPassword = envPassword
	}
	if envUrl := os.Getenv("HA_URL"); envUrl != "" {
		defaultUrl = envUrl
	}
	if envCss := os.Getenv("HA_CSS"); envCss != "" {
		defaultCss = envCss
	}
	if envPath := os.Getenv("HA_PATH"); envPath != "" {
		defaultPath = envPath
	}

	// arguments
	//
	headless = flag.Bool("headless", defaultHeadless, "Headless mode")
	username = flag.String("username", defaultUsername, "Home assistant username")
	password = flag.String("password", defaultPassword, "Home assistant password")

	restport := flag.Int("restport", defaultRestPort, "If set, startup REST server at given port")
	url := flag.String("url", defaultUrl, "Home assistant page URL")
	css := flag.String("css", defaultCss, "Home assistant CSS selector")
	path := flag.String("path", defaultPath, "Output screenshot path")

	// usage
	//
	flag.Usage = func() {
		fmt.Println("Connect to Home Assistant and take a screenshot by CSS selector")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  $HEADLESS - Headless mode")
		fmt.Println("  $HA_CSS - Home assistant CSS selector")
		fmt.Println("  $HA_PATH - Output screenshot path")
		fmt.Println("  $HA_USERNAME - Home assistant username")
		fmt.Println("  $HA_PASSWORD - Home assistant password")
		fmt.Println("  $HA_RESTPORT - If set, startup REST server at given port")
		fmt.Println("  $HA_URL - Home assistant page URL")
	}

	// parse flags
	//
	flag.Parse()

	// FIX THIS - validate

	if *restport != 0 {
		// need to start REST server
		//
		restServer := gin.Default()
		restServer.POST("", restScreenshot)
		restServer.Run(":" + strconv.Itoa(*restport))
	} else {
		// one-off run
		//
		err := screenshot(*headless, *username, *password, *url, *css, *path)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

}

func screenshot(headless bool, username string, password string, url string, css string, filename string) error {

	// setup
	//
	err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
	if err != nil {
		return fmt.Errorf("could not install playwright: %v", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not launch playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		pw.Stop()
		return fmt.Errorf("could not launch Chromium: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		browser.Close()
		pw.Stop()
		return fmt.Errorf("could not create page: %v", err)
	}

	// main page & login
	//
	log.Printf("Starting chromium at %s\n", url)
	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		browser.Close()
		pw.Stop()
		return fmt.Errorf("could not goto url: %v", err)
	}
	log.Printf("Logging in\n")
	err = page.Locator("[name=username]").First().Fill(username)
	if err != nil {
		browser.Close()
		pw.Stop()
		return fmt.Errorf("could not get username: %v", err)
	}
	err = page.Locator("[name=password]").First().Fill(password)
	if err != nil {
		browser.Close()
		pw.Stop()
		return fmt.Errorf("could not get password: %v", err)
	}
	err = page.Locator("[id=button]").Click()
	if err != nil {
		browser.Close()
		pw.Stop()
		return fmt.Errorf("could not click: %v", err)
	}

	// loop for valid screenshot - on slower machines the above wait isn't sufficient
	//
	for i := 0; i < 5; i++ {

		// wait for page to finish
		//
		page.Locator(css)
		page.WaitForURL(url, playwright.PageWaitForURLOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})

		log.Printf("Attempting screenshot %s\n", css)
		screenshot, err := page.Locator(css).Screenshot(playwright.LocatorScreenshotOptions{Path: playwright.String(filename)})
		if err != nil {
			browser.Close()
			pw.Stop()
			return fmt.Errorf("could not get screenshot: %v", err)
		}
		image, err := png.Decode(bytes.NewReader(screenshot))
		if err != nil {
			browser.Close()
			pw.Stop()
			return fmt.Errorf("could not read png: %v", err)
		}
		if isBlankImage(image) {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	log.Printf("Saved %s\n", filename)

	if err = browser.Close(); err != nil {
		return fmt.Errorf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		return fmt.Errorf("could not stop Playwright: %v", err)
	}

	return nil
}

func restScreenshot(c *gin.Context) {
	var thisRequest request

	if err := c.BindJSON(&thisRequest); err != nil {
		return
	}

	log.Printf("url=%s css=%s filename=%s\n", thisRequest.Url, thisRequest.Css, thisRequest.Filename)

	err := screenshot(*headless, *username, *password, thisRequest.Url, thisRequest.Css, thisRequest.Filename)
	if err != nil {
		c.Data(http.StatusNotFound, binding.MIMEPlain, []byte(err.Error()))
	}
}

// check to see if the screenshot is blank
//
// calculate image histogram and look for very limited colours
// (edges might have some info even if the rest of the image is blank)
func isBlankImage(image image.Image) bool {
	var histogram [16][4]int
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			r, g, b, a := image.At(x, y).RGBA()
			histogram[r>>12][0]++
			histogram[g>>12][1]++
			histogram[b>>12][2]++
			histogram[a>>12][3]++
		}
	}

	zeroCount := 0

	for _, x := range histogram {
		if x[0] == 0 && x[1] == 0 && x[2] == 0 && x[3] == 0 {
			zeroCount++
		}
	}

	return zeroCount > 10
}

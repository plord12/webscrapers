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
	"github.com/jessevdk/go-flags"
	"github.com/playwright-community/playwright-go"
)

type Options struct {
	Headless bool   `short:"e" long:"headless" description:"Headless mode" env:"HEADLESS"`
	Username string `short:"u" long:"username" description:"Home assistant username" env:"HA_USERNAME" required:"true"`
	Password string `short:"p" long:"password" description:"Home assistant password" env:"HA_PASSWORD" required:"true"`
	Restport int    `short:"r" long:"restport" description:"If set, startup REST server at given port" env:"HA_RESTPORT"`
	Url      string `short:"l" long:"url" description:"Home assistant page URL" env:"HA_URL"`
	Css      string `short:"c" long:"css" description:"Home assistant CSS selector" env:"HA_CSS"`
	Path     string `short:"a" long:"path" description:"Output screenshot path" env:"HA_PATH" default:"output.png"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)

type request struct {
	Url      string `json:"url"`
	Css      string `json:"css"`
	Filename string `json:"filename"`
}

func main() {

	// parse flags
	//
	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}

	if options.Restport != 0 {
		// need to start REST server
		//
		restServer := gin.Default()
		restServer.POST("", restScreenshot)
		restServer.Run(":" + strconv.Itoa(options.Restport))
	} else {
		if options.Url == "" || options.Css == "" || options.Path == "" {
			log.Fatal("url, css and path options must be provided when not running as a REST server")
		}
		// one-off run
		//
		err := screenshot(options.Headless, options.Username, options.Password, options.Url, options.Css, options.Path)
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
	defer pw.Stop()
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(headless)})
	if err != nil {
		return fmt.Errorf("could not launch Chromium: %v", err)
	}
	defer browser.Close()
	page, err := browser.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %v", err)
	}

	// main page & login
	//
	log.Printf("Starting chromium at %s\n", url)
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		return fmt.Errorf("could not goto url: %v", err)
	}
	log.Printf("Logging in\n")
	err = page.Locator("[name=username]").First().Fill(username)
	if err != nil {
		return fmt.Errorf("could not get username: %v", err)
	}
	err = page.Locator("[name=password]").First().Fill(password)
	if err != nil {
		return fmt.Errorf("could not get password: %v", err)
	}
	err = page.Locator("[id=button]").Click()
	if err != nil {
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
			return fmt.Errorf("could not get screenshot: %v", err)
		}
		image, err := png.Decode(bytes.NewReader(screenshot))
		if err != nil {
			return fmt.Errorf("could not read png: %v", err)
		}
		if isBlankImage(image) {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	log.Printf("Saved %s\n", filename)

	return nil
}

func restScreenshot(c *gin.Context) {
	var thisRequest request

	if err := c.BindJSON(&thisRequest); err != nil {
		return
	}

	log.Printf("url=%s css=%s filename=%s\n", thisRequest.Url, thisRequest.Css, thisRequest.Filename)

	err := screenshot(options.Headless, options.Username, options.Password, thisRequest.Url, thisRequest.Css, thisRequest.Filename)
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

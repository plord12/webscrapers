/**

Re-usuable scraping utilities

*/

package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/artdarek/go-unzip"
	"github.com/cavaliergopher/grab/v3"
	"github.com/playwright-community/playwright-go"
)

var page playwright.Page
var pw *playwright.Playwright

// Finish webscraping - check for errors and save video if needed
func Finish() {

	page.Close()

	// on error, save video if we can
	r := recover()
	if r != nil {
		log.Println("Failure:", r)
		path, err := page.Video().Path()
		if err == nil {
			log.Printf("Final screen video saved at %s\n", path)
		} else {
			log.Printf("Failed to save final video: %v\n", err)
		}
	} else {
		page.Video().Delete()
	}

}

// install Camoufox if not already installed
func installCamoufox() {
	// setup
	//
	// darwin / arm64 - https://github.com/daijro/camoufox/releases/download/v132.0-beta.15/camoufox-132.0-beta.15-mac.arm64.zip
	// linux / arm64 - https://github.com/daijro/camoufox/releases/download/v132.0-beta.15/camoufox-132.0-beta.15-lin.arm64.zip
	//

	const camoufoxVer = "132.0-beta.15"

	_, err := os.Stat("camoufox")
	if os.IsNotExist(err) {

		err := os.Mkdir("camoufox", 0750)
		if err != nil {
			panic(fmt.Sprintf("could not create directory: %v", err))
		}

		var zipFilename string
		if runtime.GOOS == "darwin" {
			zipFilename = "camoufox-" + camoufoxVer + "-mac." + runtime.GOARCH + ".zip"
		} else {
			zipFilename = "camoufox-" + camoufoxVer + "-lin." + runtime.GOARCH + ".zip"
		}

		log.Println("Installing camoufox from https://github.com/daijro/camoufox/releases/download/v" + camoufoxVer + "/" + zipFilename)

		_, err = grab.Get("camoufox", "https://github.com/daijro/camoufox/releases/download/v"+camoufoxVer+"/"+zipFilename)
		if err != nil {
			panic(fmt.Sprintf("could not download camoufox: %v", err))
		}

		uz := unzip.New("camoufox/"+zipFilename, "camoufox")
		err = uz.Extract()
		if err != nil {
			panic(fmt.Sprintf("could not unzip camoufox: %v", err))
		}
	}
}

// Start webscraping
func Start(headless *bool) playwright.Page {

	installCamoufox()

	err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		panic(fmt.Sprintf("could not install playwright: %v", err))
	}
	pw, err = playwright.Run()
	if err != nil {
		panic(fmt.Sprintf("could not launch playwright: %v", err))
	}
	defer Finish()
	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(*headless), ExecutablePath: playwright.String("camoufox/launch"), Args: []string{"--stderr", "debug.log", "--config", "{}"}})
	if err != nil {
		panic(fmt.Sprintf("could not launch Camoufox: %v", err))
	}
	page, err = browser.NewPage(playwright.BrowserNewPageOptions{RecordVideo: &playwright.RecordVideo{Dir: "videos/"}})
	if err != nil {
		panic(fmt.Sprintf("could not create page: %v", err))
	}

	return page
}

/**

Re-usuable scraping utilities

*/

package utils

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/artdarek/go-unzip"
	"github.com/cavaliergopher/grab/v3"
	stealth "github.com/jonfriesen/playwright-go-stealth"
	"github.com/playwright-community/playwright-go"
)

//go:embed launch
var launchScript string

//go:embed open
var openScript string

const camoufoxVer = "132.0-beta.15"

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/601.3.9 (KHTML, like Gecko) Version/9.0.2 Safari/601.3.9"

// Finish webscraping - check for errors and save video if needed
func Finish(page playwright.Page) {

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

// return the directory where browsers are installed
//
// same algorithm as playwright
func registryDirectory() string {

	if envPath := os.Getenv("PLAYWRIGHT_BROWSERS_PATH"); envPath != "" {
		return envPath
	}

	if runtime.GOOS == "linux" {
		if envPath := os.Getenv("XDG_CACHE_HOME"); envPath != "" {
			return path.Join(envPath, "ms-playwright")
		} else {
			return path.Join(os.Getenv("HOME"), ".cache", "ms-playwright")
		}
	} else if runtime.GOOS == "darwin" {
		return path.Join(os.Getenv("HOME"), "Library", "Caches", "ms-playwright")
	} else if runtime.GOOS == "windows" {
		if envPath := os.Getenv("LOCALAPPDATA"); envPath != "" {
			return path.Join(envPath, "ms-playwright")
		} else {
			return path.Join(os.Getenv("HOME"), "AppData", "Local")
		}
	} else {
		panic(fmt.Sprintf("unsupported operating system: %s", runtime.GOOS))
	}
}

// install Camoufox if not already installed
func installCamoufox() {

	browserDirectory := path.Join(registryDirectory(), "camoufox-"+camoufoxVer)

	_, err := os.Stat(browserDirectory)
	if os.IsNotExist(err) {

		err := os.MkdirAll(browserDirectory, 0750)
		if err != nil {
			panic(fmt.Sprintf("could not create directory: %v", err))
		}

		var zipFilename string
		if runtime.GOOS == "darwin" {
			zipFilename = "camoufox-" + camoufoxVer + "-mac." + runtime.GOARCH + ".zip"
		} else {
			zipFilename = "camoufox-" + camoufoxVer + "-lin." + runtime.GOARCH + ".zip"
		}

		// darwin / arm64 - https://github.com/daijro/camoufox/releases/download/v132.0-beta.15/camoufox-132.0-beta.15-mac.arm64.zip
		// linux / arm64 - https://github.com/daijro/camoufox/releases/download/v132.0-beta.15/camoufox-132.0-beta.15-lin.arm64.zip
		//

		log.Println("Installing camoufox from https://github.com/daijro/camoufox/releases/download/v" + camoufoxVer + "/" + zipFilename)
		log.Println("Into " + browserDirectory)

		_, err = grab.Get(browserDirectory, "https://github.com/daijro/camoufox/releases/download/v"+camoufoxVer+"/"+zipFilename)
		if err != nil {
			panic(fmt.Sprintf("could not download camoufox: %v", err))
		}

		uz := unzip.New(path.Join(browserDirectory, zipFilename), browserDirectory)
		err = uz.Extract()
		if err != nil {
			panic(fmt.Sprintf("could not unzip camoufox: %v", err))
		}

		// patch mac
		if runtime.GOOS == "darwin" {
			err = os.Rename(path.Join(browserDirectory, "launch"), path.Join(browserDirectory, "launch-orig"))
			if err != nil {
				panic(fmt.Sprintf("could not rename launch: %v", err))
			}
			if err := os.WriteFile(path.Join(browserDirectory, "launch"), []byte(launchScript), 0755); err != nil {
				panic(fmt.Sprintf("could not write new launch script: %v", err))
			}
			if err := os.WriteFile(path.Join(browserDirectory, "open"), []byte(openScript), 0755); err != nil {
				panic(fmt.Sprintf("could not write new open script: %v", err))
			}
		}
	}
}

// Start webscraping with Camoufo
func StartCamoufox(headless *bool) playwright.Page {

	installCamoufox()

	err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		panic(fmt.Sprintf("could not install playwright: %v", err))
	}
	pw, err := playwright.Run()
	if err != nil {
		panic(fmt.Sprintf("could not launch playwright: %v", err))
	}
	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(*headless), ExecutablePath: playwright.String(path.Join(registryDirectory(), "camoufox-"+camoufoxVer, "launch"))})
	if err != nil {
		panic(fmt.Sprintf("could not launch Camoufox: %v", err))
	}
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{RecordVideo: &playwright.RecordVideo{Dir: "videos/"}})
	if err != nil {
		panic(fmt.Sprintf("could not create page: %v", err))
	}

	return page
}

// Start webscraping with Chromium + stealth
func StartChromium(headless *bool) playwright.Page {

	err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
	if err != nil {
		panic(fmt.Sprintf("could not install playwright: %v", err))
	}
	pw, err := playwright.Run()
	if err != nil {
		panic(fmt.Sprintf("could not launch playwright: %v", err))
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(*headless)})
	if err != nil {
		panic(fmt.Sprintf("could not launch Chromium: %v", err))
	}

	page, err := browser.NewPage(playwright.BrowserNewPageOptions{RecordVideo: &playwright.RecordVideo{Dir: "videos/"}, UserAgent: playwright.String(userAgent)})
	if err != nil {
		panic(fmt.Sprintf("could not create page: %v", err))
	}

	// Inject stealth script
	//
	err = stealth.Inject(page)
	if err != nil {
		panic(fmt.Sprintf("could not inject stealth script: %v", err))
	}

	return page
}

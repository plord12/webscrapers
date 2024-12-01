/**

Re-usuable scraping utilities

*/

package utils

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/artdarek/go-unzip"
	"github.com/cavaliergopher/grab/v3"
	stealth "github.com/jonfriesen/playwright-go-stealth"
	"github.com/playwright-community/playwright-go"
)

const camoufoxVer = "132.0.2-beta.17"
const launchVer = "v0.0.1-alpha"

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

		var camoufoxZipFilename string
		if runtime.GOOS == "darwin" {
			camoufoxZipFilename = "camoufox-" + camoufoxVer + "-mac." + runtime.GOARCH + ".zip"
		} else {
			camoufoxZipFilename = "camoufox-" + camoufoxVer + "-lin." + runtime.GOARCH + ".zip"
		}
		launchZipFilename := "launch-" + runtime.GOOS + "-" + runtime.GOARCH + "-" + launchVer + ".zip"

		// darwin / arm64 - https://github.com/daijro/camoufox/releases/download/v132.0-beta.15/camoufox-132.0-beta.15-mac.arm64.zip
		// linux / arm64 - https://github.com/daijro/camoufox/releases/download/v132.0-beta.15/camoufox-132.0-beta.15-lin.arm64.zip
		//
		// https://github.com/plord12/webscrapers/releases/download/v0.0.1-alpha/launch-darwin-arm64-v0.0.1-alpha.zip
		//
		url := "https://github.com/daijro/camoufox/releases/download/v" + camoufoxVer + "/" + camoufoxZipFilename
		log.Println("Installing camoufox from " + url)
		log.Println("Into " + browserDirectory)
		_, err = grab.Get(browserDirectory, url)
		if err != nil {
			panic(fmt.Sprintf("could not download camoufox: %v", err))
		}
		uz := unzip.New(path.Join(browserDirectory, camoufoxZipFilename), browserDirectory)
		err = uz.Extract()
		if err != nil {
			panic(fmt.Sprintf("could not unzip camoufox: %v", err))
		}
		os.Remove(path.Join(browserDirectory, camoufoxZipFilename))

		url = "https://github.com/plord12/webscrapers/releases/download/" + launchVer + "/" + launchZipFilename
		log.Println("Installing launch from " + url)
		log.Println("Into " + browserDirectory)
		_, err = grab.Get(browserDirectory, url)
		if err != nil {
			panic(fmt.Sprintf("could not download launch: %v", err))
		}
		uz = unzip.New(path.Join(browserDirectory, launchZipFilename), browserDirectory)
		err = uz.Extract()
		if err != nil {
			panic(fmt.Sprintf("could not unzip launch: %v", err))
		}
		os.Chmod(path.Join(browserDirectory, launchZipFilename), 0755)
		os.Remove(path.Join(browserDirectory, launchZipFilename))
	}
}

// Start webscraping with Camoufo
func StartCamoufox(headless bool) playwright.Page {

	installCamoufox()

	err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		panic(fmt.Sprintf("could not install playwright: %v", err))
	}
	pw, err := playwright.Run()
	if err != nil {
		panic(fmt.Sprintf("could not launch playwright: %v", err))
	}
	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(headless), ExecutablePath: playwright.String(path.Join(registryDirectory(), "camoufox-"+camoufoxVer, "launch"))})
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
func StartChromium(headless bool) playwright.Page {

	err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
	if err != nil {
		panic(fmt.Sprintf("could not install playwright: %v", err))
	}
	pw, err := playwright.Run()
	if err != nil {
		panic(fmt.Sprintf("could not launch playwright: %v", err))
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(headless)})
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

// clean up from any previous OTP
func CleanOTP(otpCleanCommand string, otpPath string) {
	if otpCleanCommand != "" {
		log.Printf("Running %s to clean old one time password\n", otpCleanCommand)
		command := strings.Split(otpCleanCommand, " ")
		exec.Command(command[0], command[1:]...).Run()
	}
	os.Remove(otpPath)
}

// if enabled, run command to fetch OTP until it succeeds
func FetchOTP(otpCommand string) {
	if otpCommand != "" {
		log.Printf("Running %s to get one time password\n", otpCommand)
		for i := 0; i < 30; i++ {
			command := strings.Split(otpCommand, " ")
			cmd := exec.Command(command[0], command[1:]...)
			err := cmd.Run()
			if err != nil {
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
	}
}

// poll for OTP locally
func PollOTP(otpPath string) string {
	otp := ""
	for i := 0; i < 30; i++ {
		_, err := os.Stat(otpPath)
		if errors.Is(err, os.ErrNotExist) {
			time.Sleep(2 * time.Second)
		} else {
			// read otp
			//
			data, err := os.ReadFile(otpPath)
			if err == nil {
				r := regexp.MustCompile(".*([0-9][0-9][0-9][0-9][0-9][0-9]).*")
				match := r.FindStringSubmatch(string(data))
				if len(match) != 2 {
					panic(fmt.Sprintf("could not parse one time password message: %v", err))
				} else {
					otp = match[1]
				}
			}
			os.Remove(otpPath)
			break
		}
	}

	return otp
}

/**

exchange rates

*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

// currency conversions
var currencyUKP, _ = regexp.Compile(".?£([0-9.]*)[^0-9]*")
var currencyUSD, _ = regexp.Compile(".?\\$([0-9.]*)[^0-9]*")
var currencyAUD, _ = regexp.Compile(".?A\\$([0-9.]*)[^0-9]*")
var currencyEUR, _ = regexp.Compile(".?€([0-9.]*)[^0-9]*")
var currencyCAD, _ = regexp.Compile(".?CA\\$([0-9.]*)[^0-9]*")
var currencySGD, _ = regexp.Compile(".?SGD.?([0-9.]*)[^0-9]*")
var currencyARS, _ = regexp.Compile(".?ARS.?([0-9.]*)[^0-9]*")

type Rate struct {
	Code string
	Rate float64
}

var rates map[string]Rate

func getExchangeRates() {
	// should cache this
	httpClient := http.Client{}

	req, err := http.NewRequest(http.MethodGet, "https://www.floatrates.com/daily/gbp.json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create http session: %v", err)

	}
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		fmt.Fprintf(os.Stderr, "could not get exchange rates from floatrates: %v", err)
	}
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "could not read exchange rates from floatrates: %v", err)
	}
	jsonErr := json.Unmarshal(body, &rates)
	if jsonErr != nil {
		fmt.Fprintf(os.Stderr, "could not process exchange rates json: %v", err)
	}
}

func convertToGBP(currencyString string) (float64, error) {

	if currencyString == "Free" {
		return 0, nil
	}
	ukp := currencyUKP.FindStringSubmatch(currencyString)
	if len(ukp) > 0 {
		converted, err := strconv.ParseFloat(ukp[1], 32)
		return converted, err
	}
	cad := currencyCAD.FindStringSubmatch(currencyString)
	if len(cad) > 0 {
		converted, err := strconv.ParseFloat(cad[1], 32)
		if err == nil {
			return converted / rates["cad"].Rate, nil
		}
		return converted, err
	}
	usd := currencyUSD.FindStringSubmatch(currencyString)
	if len(usd) > 0 {
		converted, err := strconv.ParseFloat(usd[1], 32)
		if err == nil {
			return converted / rates["usd"].Rate, nil
		}
		return converted, err
	}
	aud := currencyAUD.FindStringSubmatch(currencyString)
	if len(aud) > 0 {
		converted, err := strconv.ParseFloat(aud[1], 32)
		if err == nil {
			return converted / rates["aud"].Rate, nil
		}
		return converted, err
	}
	eur := currencyEUR.FindStringSubmatch(currencyString)
	if len(eur) > 0 {
		converted, err := strconv.ParseFloat(eur[1], 32)
		if err == nil {
			return converted / rates["eur"].Rate, nil
		}
		return converted, err
	}
	sgd := currencySGD.FindStringSubmatch(currencyString)
	if len(sgd) > 0 {
		converted, err := strconv.ParseFloat(sgd[1], 32)
		if err == nil {
			return converted / rates["sgd"].Rate, nil
		}
		return converted, err
	}
	ars := currencyARS.FindStringSubmatch(currencyString)
	if len(ars) > 0 {
		converted, err := strconv.ParseFloat(ars[1], 32)
		if err == nil {
			return converted / rates["ars"].Rate, nil
		}
		return converted, err
	}
	return 0, errors.New("failed to parse")
}

/**

cache tests

*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func cacheAnalyse() {
	// open up cache files directly, read json into structs and process

	// look for 2 categories often used together

	// look for events with 0 categories

	folder, _ := os.UserCacheDir()
	folder = filepath.Join(folder, "events")

	entries, err := os.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	usedCategories := make(map[string]bool)
	for _, e := range entries {
		jsonData, _ := os.ReadFile(filepath.Join(folder, e.Name()))
		var record record[Cache]
		err = json.Unmarshal(jsonData, &record)
		if err != nil {
			log.Fatal(err)
		}
		for _, c := range record.Item.Categories {
			usedCategories[c] = true
		}

		if len(record.Item.Categories) == 0 {
			fmt.Printf("Event %s has zero categories\n", record.Item.Title)
		}
	}

	// look for categories not used
	for _, specifiedCategory := range append(cliOptions.Include, cliOptions.Exclude...) {
		if !usedCategories[specifiedCategory] {
			fmt.Printf("Category %s not used\n", specifiedCategory)
		}
	}
}

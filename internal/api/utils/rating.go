// rating.go

package utils

import (
	handlers "c2c/internal/api/handlers"
	models "c2c/internal/models"
	"c2c/internal/tools"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func InitRatingMap(e *handlers.Env) {

	tools.Logger.Debugf("Init rating map from: %s", e.Config.RatingMapUrl)

	var ratingMap models.RatingMap
	resp, err := http.Get(e.Config.RatingMapUrl)
	if err != nil {
		tools.Logger.Debugf("Can't get the rating map from: %s, fallback to local one", e.Config.RatingMapUrl)
		rm, err := os.ReadFile(fmt.Sprintf("%s/rating-system.json", e.Config.Installpath))
		if err != nil {
			tools.Logger.Error(err)
		} else {
			json.Unmarshal(rm, &ratingMap)
		}
	} else {
		json.NewDecoder(resp.Body).Decode(&ratingMap)

		defer resp.Body.Close()
	}

	e.RatingMap = ratingMap
}

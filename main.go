package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Sebaran struct {
	StatusCode  int          `json:"status_code"`
	SebaranData *SebaranData `json:"data"`
}

type SebaranData struct {
	StatusCode int          `json:"status_code"`
	CovidData  []*CovidData `json:"content"`
}

type CovidData struct {
	ID     string  `json:"id"`
	Kab    string  `json:"nama_kab"`
	Kec    string  `json:"nama_kec"`
	Kel    string  `json:"nama_kel"`
	Status string  `json:"status"`
	Stage  string  `json:"stage"`
	Umur   int     `json:"umur"`
	Gender string  `json:"gender"`
	Lon    float64 `json:"longitude"`
	Lat    float64 `json:"latitude"`
}

type SebaranDataMetaData struct {
	LastUpdate *time.Time `json:"last_update"`
}

type IndexerData struct {
	Kelurahan string
	Kecamatan string
	Status    map[string]int
}

var (
	IndexerJabar = make(map[string]*IndexerData)
)

func init() {
	if err := DownloadFile("jabar.json", "https://covid19-public.digitalservice.id/api/v1/sebaran/jabar"); err != nil {
		log.Panic(err)
	}
	jsonFile, err := os.Open("jabar.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	var result Sebaran

	for decoder.More() {
		decoder.Decode(&result)
	}

	// parsing data
	for _, data := range result.SebaranData.CovidData {
		if data.ID != "" && data.Status != "" {
			if val, ok := IndexerJabar[data.Kel+data.Kec]; ok {
				if _, ok := val.Status[data.Status]; ok {
					val.Status[data.Status] = val.Status[data.Status] + 1
				} else {
					val.Status[data.Status] = 1
				}
			} else {
				IndexerJabar[data.Kel+data.Kec] = &IndexerData{
					Kelurahan: data.Kel,
					Kecamatan: data.Kec,
					Status:    make(map[string]int),
				}

				IndexerJabar[data.Kel+data.Kec].Status[data.Status] = 1
			}
		}
	}
}

func DownloadFile(filepath string, url string) error {

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

func Check(kecamatan, kelurahan string) (map[string]int, bool) {
	if data, ok := IndexerJabar[kelurahan+kecamatan]; ok {
		return data.Status, ok
	} else {
		return nil, false
	}
}

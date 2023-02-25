package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type RegionTable struct {
	Version             string
	AvailableRegionsMap map[string][]string
}

type RawRegionTable struct {
	Metadata RawRegionTableMetadata `json:"metadata"`
	Prices   []RawRegionTablePrices `json:"prices"`
}

type RawRegionTableMetadata struct {
	Copyright     string `json:"copyright"`
	Disclaimer    string `json:"disclaimer"`
	FormatVersion string `json:"format:version"`
	SourceVersion string `json:"source:version"`
}

type RawRegionTablePrices struct {
	Id         string                         `json:"id"`
	Attributes RawRegionTablePricesAttributes `json:"attributes"`
}

type RawRegionTablePricesAttributes struct {
	Region      string `json:"aws:region"`
	ServiceName string `json:"aws:serviceName"`
	ServiceUrl  string `json:"aws:serviceUrl"`
}

func GetRegionTable() (RegionTable, error) {
	url := "https://api.regional-table.region-services.aws.a2z.com/index.json"
	resp, err := http.Get(url)
	if err != nil {
		return RegionTable{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return RegionTable{}, err
	}

	if resp.StatusCode != 200 {
		return RegionTable{}, fmt.Errorf("%s. %s", resp.Status, string(data[:]))
	}

	rawRegionTable := RawRegionTable{}
	if err := json.Unmarshal(data, &rawRegionTable); err != nil {
		return RegionTable{}, err
	}

	regionTable := RegionTable{
		Version:             rawRegionTable.Metadata.SourceVersion,
		AvailableRegionsMap: make(map[string][]string),
	}

	for _, price := range rawRegionTable.Prices {
		serviceFullName := price.Attributes.ServiceName
		regionName := strings.Split(price.Id, ":")[1]

		regionTable.AvailableRegionsMap[serviceFullName] = append(regionTable.AvailableRegionsMap[serviceFullName], regionName)
	}

	return regionTable, err
}

package botocore

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sync"

	errors "github.com/pkg/errors"

	"github.com/Yunsang-Jeong/listing-all-the-aws-service-apis/internal/helper"
	"github.com/Yunsang-Jeong/listing-all-the-aws-service-apis/internal/provider"
)

type botocore struct {
	helper.GithubConfig
}

type dataSource struct {
	apiVersion string
	filename   string
	sha        string
}

type dataSchema struct {
	Verion     string                `json:"version"`
	Metadata   metadata              `json:"metadata"`
	Operations map[string]operations `json:"operations"`
}

type metadata struct {
	ServiceId       string `json:"serviceId"`
	ServiceFullName string `json:"serviceFullName"`
	EndpointPrefix  string `json:"endpointPrefix"`
}

type operations struct {
	Name string `json:"name"`
}

func NewBotocore() *botocore {
	return &botocore{
		helper.GithubConfig{
			Onwer:    "boto",
			RepoName: "botocore",
		},
	}
}

func (d *botocore) GetResultFileName() string {
	return fmt.Sprintf("botocore_%s.json", d.LatestTagName)
}

func (d *botocore) SetLatestVersion() error {
	latestTag, err := helper.GetGithubRepoLatestTag(d.Onwer, d.RepoName)
	if err != nil {
		return err
	}

	d.LatestTagName = latestTag.Name
	d.LatestTagSha = latestTag.Commit["sha"]

	return nil
}

func (d *botocore) GetServiceSchemaMap() (map[string]provider.ServiceSchema, error) {
	trees, err := helper.GetGithubRepoTrees(d.Onwer, d.RepoName, d.LatestTagSha, "botocore/data")
	if err != nil {
		return nil, err
	}

	dataSourceMap := map[string]dataSource{}

	re := regexp.MustCompile(`(?P<service>.+?)/(?P<apiVersion>.+?)/service-\d.json`)
	for _, value := range trees {
		matches := re.FindStringSubmatch(value.Path)
		if matches == nil {
			continue
		}

		service := matches[re.SubexpIndex("service")]
		apiVersion := matches[re.SubexpIndex("apiVersion")]

		if _, ok := dataSourceMap[service]; ok {
			if apiVersion < dataSourceMap[service].apiVersion {
				continue
			}
		}

		dataSourceMap[service] = dataSource{
			apiVersion: apiVersion,
			filename:   fmt.Sprintf("%s/%s", "botocore/data", matches[0]),
			sha:        value.Sha,
		}
	}

	var wg sync.WaitGroup

	serviceSchemaChan := make(chan provider.ServiceSchema, len(dataSourceMap))
	errChan := make(chan error, len(dataSourceMap))

	for _, dataSource := range dataSourceMap {
		wg.Add(1)
		go d.generateServiceSchema(&wg, serviceSchemaChan, errChan, dataSource)
	}

	wg.Wait()

	close(serviceSchemaChan)
	close(errChan)

	serviceSchemaMap := map[string]provider.ServiceSchema{}

	for serviceSchema := range serviceSchemaChan {
		serviceSchemaMap[serviceSchema.ServiceFullName] = serviceSchema
	}

	errFlag := false
	for err := range errChan {
		if err != nil {
			log.Println(err)
			errFlag = true
		}
	}

	if errFlag {
		return nil, errors.New("fail to generate service schema")
	}

	return serviceSchemaMap, nil
}

func (d *botocore) generateServiceSchema(wg *sync.WaitGroup, serviceSchemaChan chan<- provider.ServiceSchema, errChan chan<- error, dataSource dataSource) {
	defer wg.Done()

	rawdata, err := helper.GetGithubRepoBlobs(d.Onwer, d.RepoName, d.LatestTagName, dataSource.filename)
	if err != nil {
		errChan <- err
		return
	}

	dataSchema := dataSchema{}
	if err := json.Unmarshal(rawdata, &dataSchema); err != nil {
		errChan <- errors.Wrap(err, "fail to unmarshal rawdata")
		return
	}

	operations := []string{}
	for operation, _ := range dataSchema.Operations {
		operations = append(operations, operation)
	}

	serviceSchemaChan <- provider.ServiceSchema{
		APIVersion:      dataSource.apiVersion,
		ServiceId:       dataSchema.Metadata.ServiceId,
		ServiceFullName: dataSchema.Metadata.ServiceFullName,
		EndpointPrefix:  dataSchema.Metadata.EndpointPrefix,
		Operations:      operations,
	}
}

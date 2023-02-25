package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Yunsang-Jeong/listing-all-the-aws-service-apis/internal/provider"
	"github.com/Yunsang-Jeong/listing-all-the-aws-service-apis/internal/provider/botocore"
)

const resultFileLocation = "data"

func main() {
	dataProviders := []provider.DataProvider{}
	dataProviders = append(dataProviders, botocore.NewBotocore())

	for _, dataProvider := range dataProviders {
		if err := os.MkdirAll("data", os.ModePerm); err != nil {
			panic(err)
		}

		if err := dataProvider.SetLatestVersion(); err != nil {
			panic(err)
		}

		serviceSchemaMap, err := dataProvider.GetServiceSchemaMap()
		if err != nil {
			panic(err)
		}

		file := fmt.Sprintf("%s/%s", resultFileLocation, dataProvider.GetResultFileName())
		data, _ := json.MarshalIndent(serviceSchemaMap, "", " ")
		if err := os.WriteFile(file, data, 0644); err != nil {
			panic(err)
		}
	}
}

// str

//

/*

datasource
 - botocore
 - aws-sdk-go-v2

리커시브하게 뭔갈 다운로드받아야 하긴해
 - 일단 목록만 뽑아

go routine
 - 가져와서
 - 분석하고
 - 리턴

결과에 대해서 json 타입으로 분류

---

release로 배포도 필요하다
github action 활용해보자
 - 매주? 단위로 트리거 시켜서 결과 확인받기

*/

// interface g

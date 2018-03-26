package eureka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"strings"
)

type Instance struct {
	Instance *EurekaInstanceInfo `json:"instance"`
}

type EurekaInstanceInfo struct {
	InstanceId string `json:"instanceId"`
	HostName   string `json:"hostName"`
	App        string `json:"app"`
	IpAddr     string `json:"ipAddr"`

	VipAddress string `json:"vipAddress"`

	SecureVipAddress string         `json:"secureVipAddress"`
	Status           string         `json:"status"`
	Port             Port           `json:"port"`
	SecurePort       Port           `json:"securePort"`
	HomePageUrl      string         `json:"homePageUrl"`
	StatusPageUrl    string         `json:"statusPageUrl"`
	HealthCheckUrl   string         `json:"healthCheckUrl"`
	DataCenterInfo   DataCenterInfo `json:"dataCenterInfo"`
}

type Port struct {
	Dollar  int  `json:"$"`
	Enabled bool `json:"@enabled"`
}
type DataCenterInfo struct {
	Class string `json:"@class"`
	Name  string `json:"name"`
}

func RegisterEureka(eurekaInstance *Instance) (string, error) {

	client2 := &http.Client{}
	marshalledInfo, err := json.Marshal(eurekaInstance)
	fmt.Println(string(marshalledInfo))
	if err != nil {
		log.Fatal(err)
	}
	eurekaUrlList := os.Getenv("EUREKA_CLIENT_SERVICEURL_DEFAULTZONE")

	if len(eurekaUrlList) == 0 {
		eurekaUrlList = "http://localhost:8761"
	}

	eurekaUrlResult := strings.Split(eurekaUrlList, ",")

	var resultError error
	resultError = nil

	var resultStatus string
	
	for _,eurekaUrl := range eurekaUrlResult{


		req, err := http.NewRequest("POST", eurekaUrl+"/eureka/apps/"+eurekaInstance.Instance.App, bytes.NewReader(marshalledInfo))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("DiscoveryIdentity-Name", "DefaultClient")
		req.Header.Add("DiscoveryIdentity-Version", "1.4")
		req.Header.Add("DiscoveryIdentity-Id", eurekaInstance.Instance.HostName)
		if err != nil {
			log.Fatal(err)
		}
	
		response, err := client2.Do(req)
		if err != nil {
			fmt.Printf(err.Error())
			resultError = err
			continue
		}
	
		bytesValue, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf(err.Error())
			resultError = err
			continue
		}
		fmt.Println(string(bytesValue))
		fmt.Println(response.Status)
		resultStatus = response.Status
	}
	

	//done := make(chan bool)
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for t := range ticker.C {
			fmt.Println("Send HearBeat at", t)
			for _,eurekaUrl := range eurekaUrlResult{
				SendHeartbeat(client2, eurekaUrl+"/eureka/apps/"+eurekaInstance.Instance.App+"/"+eurekaInstance.Instance.InstanceId)
			}
		}
	}()

	return resultStatus, resultError
}

func SendHeartbeat(client *http.Client, url string) {

	req, err := http.NewRequest("PUT", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("DiscoveryIdentity-Name", "DefaultClient")
	req.Header.Add("DiscoveryIdentity-Version", "1.4")
	req.Header.Add("DiscoveryIdentity-Id", "127.0.0.1")
	response, err := client.Do(req)

	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	bytesValue, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Println(string(bytesValue))
	fmt.Println(response.Status)
}

func GetEurekaHealth() ([]byte, error) {
	eurekaUrlList := os.Getenv("EUREKA_CLIENT_SERVICEURL_DEFAULTZONE")

	if len(eurekaUrlList) == 0 {
		eurekaUrlList = "http://localhost:8761"
	}
	eurekaUrlResult := strings.Split(eurekaUrlList, ",")

	var resultError error
	resultError = nil


	for _,eurekaUrl := range eurekaUrlResult{

		client := &http.Client{}
		req, err := http.NewRequest("GET", eurekaUrl+"/health", nil)
		resultError = err
		response, err := client.Do(req)
		if err != nil {
			log.Fatalln(err.Error())
			return nil, err
		}
		bytesValue, err := ioutil.ReadAll(response.Body)
		
		if err != nil {
			log.Fatalln(err.Error())
			return nil, err
		}
		return bytesValue, nil
	}

	return nil, resultError

	
}

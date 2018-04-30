package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

// Repository Information
type Repository struct {
	commentCount    int
	dateCreated     string
	description     string
	dockerfile      string
	fullDescription string
	isOfficial      bool
	isPrivate       bool
	isTrusted       bool
	name            string
	namespace       string
	owner           string
	repoName        string
	repoURL         string
	starCount       int
	status          string
}

//PushData Information
type PushData struct {
	pushedAt string
	pusher   string
	tag      string
	images   []string
}

//NotificationReq Information
type NotificationReq struct {
	callbackURL string
	pushData    *PushData
	repository  *Repository
}

//Container Information
type Container struct {
	ID      string
	Image   string
	ImageID string
	Labels  map[string]string
	Created int64
}

//MainfestRequest for Input
type MainfestRequest struct {
	Types   string `json:"type"`
	RepoKey string `json:"repoKey"`
	Path    string `json:"path"`
}

//JfrogInfo for Jfroginfo
type JfrogInfo struct {
	RepositoryPath string `json:"repositoryPath"`
	Created        string `json:"created"`
	LastModified   string `json:"lastModified"`
}

//JfrogRepository for JfrogRepository info
type JfrogRepository struct {
	Types string    `json:"type"`
	Info  JfrogInfo `json:"info"`
}

//JfrogImageRepository for JfrogImageRepository info
type JfrogImageRepository struct {
	Repositories []string `json:"repositories"`
}

//JfrogImageRepositoryTagList for JfrogImageRepositoryTagList info
type JfrogImageRepositoryTagList struct {
	Name  string   `json:"name"`
	Tages []string `json:"tags"`
}

//UpdateData for Update data
type UpdateData struct {
	ImageName        string    `json:"imageName"`
	Tag              string    `json:"tag"`
	LatestUpdateDate time.Time `json:"latestUpdateDate"`
	LabelName        string    `json:"labelName"`
}

//Config
type environmentInformation struct {
	JfrogURI          string `yaml:"JfrogURI"`
	Username          string `yaml:"JfrogUsername"`
	Password          string `yaml:"JfrogPassword"`
	Pattern           string `yaml:"JfrogPattern"`
	isSecure          bool   `yaml:"JfrogisSecure"`
	JfrogRepositoryUI string `yaml:"JfrogRepositoryUI"`
}

var environmentInfo environmentInformation

func (c *environmentInformation) getConf() *environmentInformation {

	yamlFile, err := ioutil.ReadFile("environment.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

var finalupdateListData []UpdateData

func checkforUpdate(res http.ResponseWriter, req *http.Request) {
	log.Println("Get Check for Update method started ")

	url := environmentInfo.JfrogURI + "_catalog"
	var updateListData []UpdateData
	log.Println(url)
	var updateData UpdateData
	body, err := ServiceRequest(url, "GET", nil)
	var jfrogImageRepository JfrogImageRepository
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &jfrogImageRepository)
	if err != nil {
		panic(err)
	}
	log.Println(jfrogImageRepository)

	for _, repo := range jfrogImageRepository.Repositories {
		if strings.Contains(repo, environmentInfo.Pattern) {
			finalTag, finalTime := getImageTags(repo)
			updateData.ImageName = repo
			updateData.Tag = finalTag
			updateData.LatestUpdateDate = finalTime
			updateListData = append(updateListData, updateData)
		} else {
			log.Println(repo)
			log.Println(environmentInfo.Pattern)
		}
	}

	response, err := json.Marshal(updateListData)
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "application/json")
	res.Write(response)
	log.Println("Get Check for Update method completed")
}

func getImageTags(imageName string) (finalTag string, finalTime time.Time) {
	log.Println("Get Image Tags  method started ")

	url := environmentInfo.JfrogURI + imageName + "/tags/list"

	body, err := ServiceRequest(url, "GET", nil)
	var jfrogImageRepositoryTagList JfrogImageRepositoryTagList
	err = json.Unmarshal(body, &jfrogImageRepositoryTagList)
	if err != nil {
		panic(err)
	}
	for _, imageTag := range jfrogImageRepositoryTagList.Tages {
		latestDate, err := getManifestImage(imageName + "/" + imageTag + "/manifest.json")
		if err == nil {
			finalTag, finalTime = Swap(finalTag, finalTime, imageName+"/"+imageTag, latestDate)
		}
	}
	log.Println("Get Image Tags  method completed ")
	return finalTag, finalTime
}

//Swap function for  Identify the latest Tag
func Swap(currentTag string, currentTime time.Time, latestTag string, latestTime time.Time) (finalTag string, finalTime time.Time) {
	if len(currentTag) > 0 {
		finalTime = currentTime
		finalTag = currentTag
	} else {
		if currentTime.Before(latestTime) {
			finalTime = latestTime
			finalTag = latestTag
		} else {
			finalTime = currentTime
			finalTag = currentTag
		}
	}
	return finalTag, finalTime
}

func getManifestImage(path string) (latestDate time.Time, err1 error) {
	log.Println("Get Image manifest  method started ")

	mainfestRequest := MainfestRequest{"file", "docker-local", path}
	var jfrogRepository JfrogRepository
	var jsonStr, err = json.Marshal(mainfestRequest)
	if err != nil {
		panic(err)
	}
	body, err := ServiceRequest(environmentInfo.JfrogRepositoryUI, "POST", jsonStr)
	err = json.Unmarshal(body, &jfrogRepository)
	if err != nil {
		panic(err)
	}
	log.Println("Get Image manifest  method completed ")
	return dateTimeConverter(jfrogRepository.Info.Created, jfrogRepository.Info.LastModified)
}

// ServiceRequest to serve http request
func ServiceRequest(URL string, Method string, input []byte) (output []byte, err error) {

	var req *http.Request
	req, err = http.NewRequest(Method, URL, bytes.NewBuffer(input))
	req.Header.Set("Content-Type", "application/json")
	if environmentInfo.isSecure == true {
		req.SetBasicAuth(environmentInfo.Username, environmentInfo.Password)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return body, err
}

func getManifest(res http.ResponseWriter, req *http.Request) {
	log.Println("Get Manifest method started ")

	var mainfestRequest MainfestRequest
	var jfrogRepository JfrogRepository

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&mainfestRequest); err == io.EOF {

	} else if err != nil {
		log.Fatal(err)
	}
	mainfestRequestBody, err := json.Marshal(mainfestRequest)
	var jsonStr = []byte(mainfestRequestBody)
	req, err = http.NewRequest("POST", environmentInfo.JfrogRepositoryUI, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "P@ssw0rd@123")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err = json.Unmarshal(body, &jfrogRepository)

	log.Println("Get Manifest method completed")
	res.Header().Set("Content-Type", "application/json")
	res.Write(body)
}

func test(res http.ResponseWriter, req *http.Request) {
	log.Println("Get Manifest method started ")
	var jfrogImageRepository JfrogImageRepository
	url := environmentInfo.JfrogURI + "_catalog"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "P@ssw0rd@123")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err = json.Unmarshal(body, &jfrogImageRepository)

	log.Println(jfrogImageRepository)

	log.Println("Get Manifest method completed")
	res.Header().Set("Content-Type", "application/json")
	res.Write(body)
}

func dateTimeConverter(createdDateString string, modifiedDateString string) (latestDate time.Time, err error) {
	var t3 time.Time
	log.Println("Date Time Converter method started ")
	if len(createdDateString) > 0 && len(modifiedDateString) > 0 {

		created := []rune(createdDateString)
		lastModified := []rune(modifiedDateString)
		createddate := "20" + string(created[6:8]) + "-" + string(created[3:5]) + "-" + string(created[0:2]) + "T" + string(created[9:17]) + ".0Z"
		lastModifieddate := "20" + string(lastModified[6:8]) + "-" + string(lastModified[3:5]) + "-" + string(lastModified[0:2]) + "T" + string(lastModified[9:17]) + ".0Z"

		t1, err := time.Parse(time.RFC3339, createddate)

		t2, err := time.Parse(time.RFC3339, lastModifieddate)

		if t1.Before(t2) {
			t3 = t2
		} else {
			t3 = t1
		}
		log.Println(err)

	}
	log.Println("Date Time Converter method completed ")
	return t3, err

}

func main() {

	environmentInfo.getConf()

	router := mux.NewRouter()
	router.HandleFunc("/getManifest", getManifest).Methods("POST")
	router.HandleFunc("/checkforUpdate", checkforUpdate).Methods("GET")
	router.HandleFunc("/test", test).Methods("GET")

	log.Fatal(http.ListenAndServe(":8003", router))
}

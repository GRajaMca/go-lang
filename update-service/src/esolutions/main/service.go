package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var finalupdateListData []UpdateData

func processUpdate(res http.ResponseWriter, req *http.Request) {
	log.Println("Get Process for Update method started ")
	var updateData []UpdateData
	cli := getClientInstance()
	var response []string
	if req.Body == nil {
		http.Error(res, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(req.Body).Decode(&updateData)

	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range updateData {
		serviceInfo := getServiceInfo(cli, repo.LabelName)
		updateService(cli, repo.LabelName, serviceInfo, repo.ImageName+":"+repo.Tag)
		err := removeService(cli, repo.LabelName)
		serviceResponse, err := createService(cli, serviceInfo, repo.ImageName+":"+repo.Tag)
		if err != nil {
			panic(err)
		}
		response = append(response, serviceResponse.ID)
	}

	result, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "application/json")
	res.Write(result)
	log.Println("Get Process for Update method completed ")
}

func updateDockerService(res http.ResponseWriter, req *http.Request) {
	log.Println("Get updateDockerService method started ")
	var updateData []UpdateData
	cli := getClientInstance()
	var response []string
	if req.Body == nil {
		http.Error(res, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(req.Body).Decode(&updateData)

	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range updateData {
		serviceInfo := getServiceInfo(cli, repo.LabelName)
		serviceUpdateResponse, err := updateService(cli, repo.LabelName, serviceInfo, repo.ImageName+":"+repo.Tag)
		if err != nil {
			panic(err)
		}
		response = serviceUpdateResponse.Warnings
	}

	result, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "application/json")
	res.Write(result)
	log.Println("Get updateDockerService method completed ")
}

func checkforUpdate(res http.ResponseWriter, req *http.Request) {
	log.Println("Get Check for Update method started ")

	url := environmentInfo.JfrogURI + "_catalog"
	var updateListData []UpdateData

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

	response, err := json.Marshal(getUpdateInformation(updateListData))
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "application/json")
	res.Write(response)
	log.Println("Get Check for Update method completed")
}

func getUpdateInformation(updateListData []UpdateData) (finalupdateListData []UpdateData) {
	cli := getClientInstance()
	log.Println("Get Update Information method started ")
	containerInformation := getContainerList(cli)
	log.Println(updateListData)
	for _, containerInfo := range containerInformation {
		for _, currentUpdateData := range updateListData {
			log.Println(containerInfo.Image)
			log.Println(currentUpdateData.ImageName)
			if containerInfo.Image == currentUpdateData.ImageName {
				i, err := strconv.ParseInt(string(containerInfo.Created), 10, 64)
				if err != nil {
					panic(err)
				}
				tm := time.Unix(i, 0)
				if tm.Before(currentUpdateData.LatestUpdateDate) {
					var finalupdateData UpdateData
					finalupdateData.ImageName = currentUpdateData.ImageName
					finalupdateData.Tag = currentUpdateData.Tag
					finalupdateData.LatestUpdateDate = currentUpdateData.LatestUpdateDate
					finalupdateData.LabelName = containerInfo.Labels["com.docker.swarm.service.name"]
					finalupdateListData = append(finalupdateListData, finalupdateData)
				} else {
					log.Println("Condition failed ")
				}
			}
		}

	}
	log.Println("Get Update Information method completed  ")
	return finalupdateListData

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
			finalTag, finalTime = Swap(finalTag, finalTime, imageName+":"+imageTag, latestDate)
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
	req.SetBasicAuth(environmentInfo.Username, environmentInfo.Password)
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

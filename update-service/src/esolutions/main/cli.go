package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

var notificatioRequest NotificationReq

var containerInformation []Container

var serviceInfo swarm.Service

var serviceResponse types.ServiceCreateResponse

//
func getClientInstance() *client.Client {
	log.Println("Get Client Instance method execution started")
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	log.Println("Get Client Instance method execution completed")
	return cli
}

func getContainerList(cli *client.Client) []Container {
	log.Println("Get ContainerList method execution started")

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		var containerInfo Container
		containerInfo.ID = container.ID
		containerInfo.Image = container.Image
		containerInfo.ImageID = container.ImageID
		containerInfo.Labels = container.Labels
		containerInfo.Created = container.Created
		containerInformation = append(containerInformation, containerInfo)
	}
	log.Println("Get ContainerList method execution completed")
	return containerInformation
}

func getServiceInfo(cli *client.Client, serviceID string) swarm.Service {
	log.Println("Get Service Info method execution started")
	service, byte, err := cli.ServiceInspectWithRaw(context.Background(), serviceID, types.ServiceInspectOptions{})
	if err != nil {
		log.Println(byte)
		panic(err)
	}
	log.Println("Get Service Info method execution completed")
	return service
}
func updateService(cli *client.Client, serviceID string, service swarm.Service, repositoryURL string) (types.ServiceUpdateResponse, error) {
	log.Println("Update Service method execution started")
	res, err := cli.ServiceUpdate(context.Background(), serviceID, swarm.Version{}, service.Spec, types.ServiceUpdateOptions{})
	log.Println("Update Service method execution completed")
	return res, err
}

func removeService(cli *client.Client, serviceName string) error {
	log.Println("Remove Service method execution started / Completed")
	return cli.ServiceRemove(context.Background(), serviceName)
}

func createService(cli *client.Client, serviceInfo swarm.Service, repositoryURL string) (types.ServiceCreateResponse, error) {
	log.Println("Create Service method execution started")
	serviceInfo.Spec.TaskTemplate.ContainerSpec.Image = repositoryURL
	log.Println("Create Service method execution completed")
	return cli.ServiceCreate(context.Background(), serviceInfo.Spec, types.ServiceCreateOptions{})
}

func getNotification(res http.ResponseWriter, req *http.Request) {
	cli := getClientInstance()
	log.Println("Get Notification method started ")
	containerInformation := getContainerList(cli)

	for _, containerInfo := range containerInformation {
		serviceInfo := getServiceInfo(cli, containerInfo.Labels["com.docker.swarm.service.name"])
		error := removeService(cli, containerInfo.Labels["com.docker.swarm.service.name"])
		params := mux.Vars(req)
		serviceResponse, error = createService(cli, serviceInfo, params["RepositoryURL"]+params["ImageURL"])
		if error != nil {
			log.Println("Get Notification method completed")
			panic(error)
		} else {
			log.Println("Get Notification method completed")
			json.NewEncoder(res).Encode(serviceResponse)
		}
	}

}

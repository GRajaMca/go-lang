package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func (c *environmentInformation) getConf() *environmentInformation {

	yamlFile, err := ioutil.ReadFile("environment.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	if len(os.Getenv("JfrogURI")) > 0 {
		c.JfrogURI = os.Getenv("JfrogURI")
	}

	if len(os.Getenv("JfrogUsername")) > 0 {
		c.Username = os.Getenv("JfrogUsername")
	}
	if len(os.Getenv("JfrogPassword")) > 0 {
		c.Password = os.Getenv("JfrogPassword")
	}

	if len(os.Getenv("JfrogPattern")) > 0 {
		c.Pattern = os.Getenv("JfrogPattern")
	}

	if len(os.Getenv("JfrogRepositoryUI")) > 0 {
		c.JfrogRepositoryUI = os.Getenv("JfrogRepositoryUI")
	}
	secure := os.Getenv("JfrogisSecure")
	if len(secure) > 0 && strings.Contains(secure, "false") {
		c.isSecure = true
	}
	return c
}

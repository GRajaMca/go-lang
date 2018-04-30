package main

import "time"

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

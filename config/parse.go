package config

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	logger "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

type GitlabElement struct {
	Name               string         `yaml:"name"`
	Namespace          string         `yaml:"namespace"`
	NamespaceOld       string         `yaml:"namespace_old"`
	State              string         `yaml:"state"`
	Description        string         `yaml:"description"`
	Visibility         string         `yaml:"visibility,omitempty"`
	Avatar             string         `yaml:"avatar,omitempty"`
	CleanUnmanagedVars bool           `yaml:"clean_unmanaged_variables"`
	CIConfigPath       string         `yaml:"ci_config_path,omitempty"`
	Sched              []Sched        `yaml:"sched,omitempty"`
	VariablesFile      string         `yaml:"variables_file,omitempty"`
	Variables          []Variable     `yaml:"variables,omitempty"`
	DeployFreezes      []DeployFreeze `yaml:"deploy_freeze,omitempty"`
	Hooks              []Hook         `yaml:"webhooks,omitempty"`
}

type DeployFreeze struct {
	FreezeStart  string `yaml:"freeze_start"`
	FreezeEnd    string `yaml:"freeze_end"`
	CronTimezone string `yaml:"cron_timezone"`
}

type Sched struct {
	Ref         string     `yaml:"ref"`
	Description string     `yaml:"description"`
	Cron        string     `yaml:"cron"`
	Variables   []Variable `yaml:"variables,omitempty"`
}

type Variable struct {
	Key          string `yaml:"key"`
	State        string `yaml:"state,omitempty"`
	VariableType string `yaml:"variable_type"`
	Protected    bool   `yaml:"protected,omitempty"`
	Masked       bool   `yaml:"masked,omitempty"`
	Environment  string `yaml:"environment,omitempty"`
	Value        string `yaml:"value"`
}

type FileVariables struct {
	Variables []Variable `yaml:"variables,omitempty"`
}

type Hook struct {
	ConfidentialIssuesEvents bool   `yaml:"confidential_issues_events,omitempty"`
	ConfidentialNoteEvents   bool   `yaml:"condidential_note_events,omitempty"`
	DeploymentEvents         bool   `yaml:"deployment_events,omitempty"`
	EnableSSLVerification    bool   `yaml:"enable_ssl_verification,omitempty"`
	IssuesEvents             bool   `yaml:"issues_events,omitempty"`
	JobEvents                bool   `yaml:"job_events,omitempty"`
	MergeRequestsEvents      bool   `yaml:"merge_requests_evenets,omitempty"`
	NoteEvents               bool   `yaml:"note_events,omitempty"`
	PipelineEvents           bool   `yaml:"pipeline_events,omitempty"`
	PushEvents               bool   `yaml:"push_evenets,omitempty"`
	PushEventsBranchFilter   string `yaml:"push_events_branch_filter,omitempty"`
	ReleasesEvents           bool   `yaml:"releases_events,omitempty"`
	TagPushEvents            bool   `yaml:"tag_push_events,omitempty"`
	WikiPageEvents           bool   `yaml:"wiki_page_events,omitempty"`
	Token                    string `yaml:"token,omitempty"`
	URL                      string `yaml:"url,omitempty"`
}

const (
	ymlExt  = ".yml"
	yamlExt = ".yaml"
)

type GACFile struct {
	Groups   []GitlabElement `yaml:"groups"`
	Projects []GitlabElement `yaml:"projects"`
}

func readFile(root string, file fs.FileInfo) ([]byte, error) {
	if e := filepath.Ext(file.Name()); e == ymlExt || e == yamlExt {
		yamlFile, err := os.Open(filepath.Join(root, file.Name()))
		if err != nil {
			return nil, err
		}
		defer yamlFile.Close()

		return io.ReadAll(yamlFile)
	}
	return nil, nil
}

func unmarshal(data []byte) (*GACFile, error) {
	var gac GACFile
	if err := yaml.Unmarshal(data, &gac); err != nil {
		return nil, err
	}
	return &gac, nil
}

func ParseYaml(rootDir string) ([]GitlabElement, []GitlabElement, error) {
	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":   err,
			"ReadDir": rootDir,
		}).Error("Error occured")
		return nil, nil, err
	}

	var groups []GitlabElement
	var projects []GitlabElement

	for _, file := range files {
		data, err := readFile(rootDir, file)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"File":  file,
			}).Error("Error occured")
			continue
		}

		gac, err := unmarshal(data)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"File":  file,
			}).Error("Error occured")
			continue
		}

		groups = append(groups, gac.Groups...)
		projects = append(projects, gac.Projects...)
	}

	return groups, projects, nil
}

func ParseVariableFile(filePath string) (FileVariables, error) {
	var fileVariables FileVariables
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fileVariables, err
	}

	err = yaml.Unmarshal(fileBytes, &fileVariables)
	if err != nil {
		return fileVariables, err
	}
	return fileVariables, nil
}

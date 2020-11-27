package commands

import (
	"strings"
	"time"

	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"

	searchutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)

// GetRemoteCounterCommand Get remote counter command
func GetRemoteCounterCommand() components.Command {
	return components.Command{
		Name:        "remote-counter",
		Description: "Count remote downloads.",
		Aliases:     []string{"rc"},
		Arguments:   []components.Argument{},
		Flags:       getFlags(),
		EnvVars:     []components.EnvVar{},
		Action: func(c *components.Context) error {
			return remoteCounterCmd(c)
		},
	}
}

func getFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:         "user",
			Description:  "Users to count downloads for.",
			DefaultValue: "",
		},
		components.StringFlag{
			Name:         "repo",
			Description:  "Repositories to count downloads for.",
			DefaultValue: "",
		},
		components.StringFlag{
			Name:         "before",
			Description:  "End date limit to count downloads.",
			DefaultValue: "",
		},
		components.StringFlag{
			Name:         "after",
			Description:  "Start date limit to count downloads.",
			DefaultValue: "",
		},
		components.StringFlag{
			Name:         "export",
			Description:  "Export type (csv, json).",
			DefaultValue: "",
		},
	}
}

type remoteCounterConfiguration struct {
	users  []string
	repos  []string
	before time.Time
	after  time.Time
	out    string
}

func remoteCounterCmd(c *components.Context) error {
	var conf = new(remoteCounterConfiguration)
	conf.users = parseCsvAsArray(c.GetStringFlagValue("user"))
	conf.repos = parseCsvAsArray(c.GetStringFlagValue("repo"))
	conf.before = parseStringAsTimestamp(c.GetStringFlagValue("before"))
	conf.after = parseStringAsTimestamp(c.GetStringFlagValue("after"))
	conf.out = parseExportType(c.GetStringFlagValue("export"))

	log.Output(doCount(conf))

	return nil
}

func parseCsvAsArray(csv string) []string {
	return strings.Split(csv, ",")
}

func parseStringAsTimestamp(str string) time.Time {
	t, _ := time.Parse(time.RFC3339, str)

	return t
}

func parseExportType(str string) string {
	switch str {
	case "csv":
	case "json":
		return str
	}

	return ""
}

func doCount(c *remoteCounterConfiguration) string {
	var localRepoCounterKey = "repo-counter-local"

	// autenticate
	artDetails, err := config.GetDefaultArtifactoryConf()
	if err != nil {
		log.Error("error while reading configuration", err)
		return ""
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		log.Error("error while authenticating", err)
		return ""
	}

	// assert repository existence
	utils.CheckIfRepoExists(localRepoCounterKey, artAuth)

	// create configuration
	rtConf := new(searchutils.CommonConfImpl)
	rtConf.SetArtifactoryDetails(artAuth)

	// collect files
	for _, user := range c.users {
		for _, repo := range c.repos {
			aqlQuery := buildAQL(c, localRepoCounterKey, user, repo)
			resultReader, err := searchutils.ExecAqlSaveToFile(aqlQuery, rtConf)
			if err != nil {
				log.Error("error while processing count", err)
				return ""
			}
			defer resultReader.Close()
		}
	}

	// aggregate counter

	// prepare output

	return ""
}

func buildAQL(c *remoteCounterConfiguration, localRepoCounterKey string, user string, repo string) (aqlQuery string) {
	aqlQuery = `items.find({` +
		`"type":"file",` +
		`"repo":` + localRepoCounterKey + "," +
		`"created" : {"$gt" :` + c.after.Format("2006-01-02T15:04:05") + "}," +
		`"created" : {"$lt" :` + c.before.Format("2006-01-02T15:04:05") + "}," +
		`"item.path : {"$eq" :` + user + "/" + repo + "}" +
		`})`

	return aqlQuery
}

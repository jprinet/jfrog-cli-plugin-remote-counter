package commands

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	searchutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const localRepoCounterKey = "remote-counter-local"
const aliasAllUsers = "ALL_USERS"
const aliasAllRepositories = "ALL_REPOS"
const githubLink = "https://github.com/jprinet/jfrog-cli-plugin-remote-counter"

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
			Description:  "csv list of users to count downloads for (all by default).",
			DefaultValue: "*",
		},
		components.StringFlag{
			Name:         "repo",
			Description:  "csv list of repositories to count downloads for (all by default).",
			DefaultValue: "*",
		},
		components.StringFlag{
			Name:         "after",
			Description:  "Start date to count downloads (formatted as 2006-01-02 or 2006-01-02T15:04:05, 1970-12-31 by default).",
			DefaultValue: "1970-12-31",
		},
		components.StringFlag{
			Name:         "before",
			Description:  "End date to count downloads (formatted as 2006-01-02 or 2006-01-02T15:04:05, 2999-12-31 by default).",
			DefaultValue: "2999-12-31",
		},
		components.StringFlag{
			Name:         "csv",
			Description:  "output csv filename.",
			DefaultValue: "",
		},
		components.StringFlag{
			Name:         "server-id",
			Description:  "Artifactory server ID configured using the config command. If not specified, the default configured Artifactory server is used.",
			DefaultValue: "",
		}}
}

type remoteCounterConfiguration struct {
	users    []string
	repos    []string
	after    time.Time
	before   time.Time
	csv      string
	serverID string
}

func remoteCounterCmd(c *components.Context) error {
	// parse configuration
	var conf = new(remoteCounterConfiguration)

	conf.users = parseCsvAsArray(c.GetStringFlagValue("user"))
	conf.repos = parseCsvAsArray(c.GetStringFlagValue("repo"))

	var err error
	conf.after, err = parseStringAsTimestamp(c.GetStringFlagValue("after"))
	if err != nil {
		return err
	}
	conf.before, err = parseStringAsTimestamp(c.GetStringFlagValue("before"))
	if err != nil {
		return err
	}

	conf.csv = c.GetStringFlagValue("csv")
	conf.serverID = c.GetStringFlagValue("server-id")

	// run command
	processError := doCount(conf)
	if processError != nil {
		log.Error("error while processing", processError)
	}

	return nil
}

func parseCsvAsArray(csv string) []string {
	return strings.Split(csv, ",")
}

func parseStringAsTimestamp(str string) (time.Time, error) {
	var formattedStr string
	if len(str) == len("2006-01-02") {
		formattedStr = str + "T00:00:00.000000000Z"
	} else if len(str) >= len("2006-01-02T15:04:05") {
		formattedStr = str[:len("2006-01-02T15:04:05")] + ".000000000Z"
	} else {
		return time.Time{}, errors.New("Input date format not supported (please use 2006-01-02 or 2006-01-02T15:04:05): " + str)
	}

	t, err := time.Parse(time.RFC3339Nano, formattedStr)
	if err != nil {
		return time.Time{}, errors.New("error parsing input date")
	}

	return t, nil
}

func doCount(c *remoteCounterConfiguration) error {
	// connect to Artifactory
	rtConf, connectionError := rtConnect(c.serverID)
	if connectionError != nil {
		log.Error(connectionError)
		return errors.New("error connecting to Artifactory")
	}

	// assert counter repository existence
	counterRepoCheckError := utils.CheckIfRepoExists(localRepoCounterKey, rtConf.GetArtifactoryDetails())
	if counterRepoCheckError != nil {
		return errors.New("Local repository " + localRepoCounterKey + " not found\nAre you sure the remote-counter user plugin is installed on your instance?\nPlease check the prerequisistes section:\n" + githubLink)
	}

	// prepare output file
	var writer *csv.Writer
	if c.csv != "" {
		log.Info("Output to csv file " + c.csv)

		file, err := os.Create(c.csv)
		if err != nil {
			log.Error(err)
			return errors.New("error creating output file")
		}
		defer file.Close()
		writer = csv.NewWriter(file)
		defer writer.Flush()
	}

	// loop over users
	var totalCounter = 0
	for _, user := range c.users {
		// loop over repositories
		var totalCounterForUser = 0
		for _, repo := range c.repos {
			// collect files
			aqlQuery := buildAQL(c, localRepoCounterKey, user, repo)
			resultReader, err := searchutils.ExecAqlSaveToFile(aqlQuery, rtConf)
			if err != nil {
				log.Error(err)
				return errors.New("Error running AQL query")
			}

			// Iterate over the results.
			var counter = 0
			for currentResult := new(searchutils.ResultItem); resultReader.NextRecord(currentResult) == nil; currentResult = new(searchutils.ResultItem) {
				log.Debug("Found artifact: " + currentResult.Name + "(" + currentResult.Path + ")")
				counter++
			}
			if err := resultReader.GetError(); err != nil {
				log.Error(err)
				return errors.New("Unable to read results")
			}
			if repo != "*" {
				logOutput(getCurrentUser(user), repo, counter, writer)
			}

			// aggregate counter per user
			totalCounterForUser += counter

			defer resultReader.Close()
			resultReader.Reset()
		}

		// aggregate counter for all users
		totalCounter += totalCounterForUser

		if user != "*" {
			logOutput(user, aliasAllRepositories, totalCounterForUser, writer)
		}
	}

	// aggregate counter for all users
	logOutput(aliasAllUsers, aliasAllRepositories, totalCounter, writer)

	return nil
}

func getCurrentUser(user string) string {
	var currentUser string
	if user != "*" {
		currentUser = user
	} else {
		currentUser = aliasAllUsers
	}
	return currentUser
}

func logOutput(user string, repo string, counter int, writer *csv.Writer) error {
	log.Info(user + "," + repo + "," + strconv.Itoa(counter))
	if writer != nil {
		err := writer.Write([]string{user, repo, strconv.Itoa(counter)})
		if err != nil {
			log.Warn("Unable to write to file", err)
		}
	}

	return nil
}

func rtConnect(serverID string) (searchutils.CommonConf, error) {
	artDetails, err := commands.GetConfig(serverID, false)
	if err != nil {
		return nil, err
	}
	if artDetails.Url == "" {
		return nil, errors.New("no server-id was found, or the server-id has no url")
	}
	artDetails.Url = clientutils.AddTrailingSlashIfNeeded(artDetails.Url)
	err = config.CreateInitialRefreshableTokensIfNeeded(artDetails)
	if err != nil {
		return nil, err
	}

	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}

	rtConf := new(searchutils.CommonConfImpl)
	rtConf.SetArtifactoryDetails(artAuth)

	log.Info("Connected to " + rtConf.GetArtifactoryDetails().GetUrl())

	return rtConf, nil
}

func buildAQL(c *remoteCounterConfiguration, localRepoCounterKey string, user string, repo string) (aqlQuery string) {
	log.Debug("after = " + c.after.String())
	log.Debug("before = " + c.before.String())
	log.Debug("user = " + c.before.String())
	log.Debug("repo = " + c.before.String())

	aqlQuery = `items.find({` +
		`"type":"file",` +
		`"repo":%q,` +
		`"created" : {"$gt" :%q},` +
		`"created" : {"$lt" :%q},` +
		`"path" : {"$match" :%q}` +
		`})`

	return fmt.Sprintf(aqlQuery, localRepoCounterKey, c.after.Format("2006-01-02T15:04:05"), c.before.Format("2006-01-02T15:04:05"), user+"/"+repo)
}

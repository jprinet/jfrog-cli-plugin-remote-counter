# remote-counter

## About this plugin
This plugin counts downloads initiated from Artifactory's remote repositories. 

It is possible to filter:
- per users
- per remote repositories
- on date range

Results are sent to stdout and can be exported to a csv file.

This plugin may be particularily helpful in situations where a central repository (ie. Dockerhub) is limiting the download rate.
It allows to understand the distribution of the requests per user.

## Installation with JFrog CLI
Since this plugin is currently not included in [JFrog CLI Plugins Registry](https://github.com/jfrog/jfrog-cli-plugins-reg), it needs to be built and installed manually. Follow these steps to install and use this plugin with JFrog CLI.
1. Make sure JFrog CLI is installed on you machine by running ```jfrog```. If it is not installed, [install](https://jfrog.com/getcli/) it.
2. Create a directory named ```plugins``` under ```~/.jfrog/``` if it does not exist already.
3. Clone this repository.
4. CD into the root directory of the cloned project.
5. Run ```go build``` to create the binary in the current directory.
6. Copy the binary into the ```~/.jfrog/plugins``` directory.

## Usage

### Prerequisites
As of now, Artifactory is not exposing metrics on remote download counts. 

- [Install](https://www.jfrog.com/confluence/display/JFROG/User+Plugins#UserPlugins-DeployingPlugins) the [remote-counter](artifactory-user-plugin/remote-counter.groovy) user plugin in your instance to get those computed. 

- This plugin needs a generic local repository named **remote-counter-local** to be created in your Artifactory instance and all users to be allowed to deploy into this repository.

This plugin creates an horodated empty file each time a download is happening from a remote repository (list of remote repositories can be tweaked).
There is a policy to remove files older than 30 days by default.

### Commands
* remote-counter
    - Flags:
        - user: filter downloads on given csv list of users **[Default: all]**
        - repo: filter downloads on given csv list of remote repositories **[Default: all]**
        - after: filter downloads issued after given date (formatted as 2006-01-02 or 2006-01-02T15:04:05) **[Default: 1970-12-31]**
        - before: filter downloads issued before given date (formatted as 2006-01-02 or 2006-01-02T15:04:05) **[Default: 2999-12-31]**
        - csv: export output to a csv file with the given name **[Default: none]**
    - Example:
    ```
    $ jfrog remote-counter
    [Info] Connected to http://artifactory-local.com/artifactory/
    [Info] ALL_USERS,ALL_REPOS,17379

    $ jfrog remote-counter --user=alice,bob,pipelines --repo=foo-mvn,foo-go,bar-mvn,bar-docker --before=2020-12-31T10:00:00 --after=2020-01-31T10:00:00 --csv=output.csv
    [Info] Connected to http://artifactory-local.com/artifactory/
    [Info] alice,foo-mvn,42
    [Info] alice,foo-go,0
    [Info] alice,bar-mvn,0
    [Info] alice,bar-docker,0
    [Info] alice,ALL_REPOS,42
    [Info] bob,foo-mvn,1
    [Info] bob,foo-go,2
    [Info] bob,bar-mvn,150
    [Info] bob,bar-docker,4
    [Info] bob,ALL_REPOS,157
    [Info] pipelines,foo-mvn,1679
    [Info] pipelines,foo-go,13456
    [Info] pipelines,bar-mvn,1589
    [Info] pipelines,bar-docker,456
    [Info] pipelines,ALL_REPOS,17180
    [Info] ALL_USERS,ALL_REPOS,17379

    ```

## Release Notes
The release notes are available [here](RELEASE.md).

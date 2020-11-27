# hello-frog

## About this plugin
This plugin counts downloads initiated from Artifactory's remote repositories. 
Counts are issued per user and per remote repository.

It is possible to filter:
- on user
- on remote repository
- on date range

Results are sent to stdout and can be exported to a:
- csv file 
- json file

This plugin may be particularily helpful in situations where a central repository (ie. Dockerhub) is limiting the download rate.
It allows to understand the distribution of the requests per user.

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install remote-counter`

Installing a specific version:

`$ jfrog plugin install remote-counter@version`

Uninstalling a plugin

`$ jfrog plugin uninstall remote-counter`

## Usage

### Prerequisites
As of now, Artifactory is not exposing metrics on remote download counts. [Install](https://www.jfrog.com/confluence/display/JFROG/User+Plugins#UserPlugins-DeployingPlugins) the [remote-counter](artifactory-user-plugin/remote-counter.groovy) user plugin in your instance to get those computed. 

This plugin needs the remote-counter-local generic local repository to be created in your Artifactory instance and all users to be allowed to deploy into this repository.

This plugin creates an horodated empty file each time a download is happening from a remote repository.

### Commands
* remote-counter
    - Flags:
        - user: filter downloads on given csv list of users **[Default: all]**
        - repo: filter downloads on given csv list of remote repositories **[Default: all]**
        - before: filter downloads issued before given date YYYY-MM-DDThh:mm:ss (THH:mm:ss is optional) **[Default: none]**
        - after: filter downloads issued after given date YYYY-MM-DDThh:mm:ss (THH:mm:ss is optional) **[Default: none]**
        - export: export output to a file (csv or json) **[Default: none]**
    - Example:
    ```
    $ jfrog remote-counter --user=alice,bob,pipelines --repo=foo-mvn,foo-go,bar-mvn,bar-docker --before=2020-31-12T10:00:00 --after=2020-31-01T10:00:00 --export=csv
    alice,foo-mvn,42
    alice,foo-go,0
    alice,bar-mvn,0
    alice,bar-docker,0
    bob,foo-mvn,1
    bob,foo-go,2
    bob,bar-mvn,150
    bob,bar-docker,4
    pipelines,foo-mvn,1679
    pipelines,foo-go,13456
    pipelines,bar-mvn,1589
    pipelines,bar-docker,456
    ```

## Additional info
None.

## Release Notes
The release notes are available [here](RELEASE.md).

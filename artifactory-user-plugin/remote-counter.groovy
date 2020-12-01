import org.artifactory.request.Request
import org.artifactory.repo.RepoPath
import org.artifactory.repo.RepoPathFactory
import org.artifactory.security.User
import org.artifactory.security.Security
import org.artifactory.search.Searches
import org.artifactory.search.aql.AqlResult

/**
 * This plugin creates an horodated file each time a remote download happens.
 * The file is stored in remote-counter-local/${username}.
 *
 * user needs the deploy permission in the remote-counter-local repository
 *
 * Configuration:
 * - Monitored repositories can be tweaked adding elements to MONITORED_REPOSITORIES
 * - Once a week, elements older than RETENTION_IN_DAYS (default 30 days) are deleted from remote-counter-local
 *
 */

class RemoteCounterConstants {

    static final String LOCAL_COUNTER_REPO = "remote-counter-local"

    // add elements as string to MONITORED_REPOSITORIES to whitelist repositories to monitor
    static final def MONITORED_REPOSITORIES = []

    // customize retention policy if needed
    static final int RETENTION_IN_DAYS = 30
}

download {
    
    afterRemoteDownload { request, repoPath ->
        try {
            if (RemoteCounterConstants.MONITORED_REPOSITORIES.size() == 0 || RemoteCounterConstants.MONITORED_REPOSITORIES.contains("" + repoPath.getRepoKey())) {
                log.debug("request = ${request}")
                log.debug("repoPath = ${repoPath}")

                String username = security.currentUser().getUsername()
                log.debug("user = ${username}")

                String timestamp = java.time.LocalDateTime.now()
                log.debug("timestamp = " + timestamp)

                // instantiate counter file
                RepoPath counterFile = RepoPathFactory.create(RemoteCounterConstants.LOCAL_COUNTER_REPO, username + "/" + repoPath.getRepoKey() + "/" + timestamp)

                // deploy counter file
                if (security.canDeploy(counterFile)) {
                    log.debug("Deploying $counterFile")
                    repositories.deploy(counterFile, new java.io.StringBufferInputStream(""))
                } else {
                    log.warn("Can't deploy $counterFile (user ${username} has no deploy permission)")
                }
            } else {
                log.debug("Repository " + repoPath.getRepoKey() + " not monitored")
            }
        } catch(Exception e){
            log.error("Error in remote-counter", e);
        }
    }
}

jobs {    

    // clean artifacts older than RETENTION_IN_DAYS in remote-counter-local every Monday at 20.30
    cleanup(cron: "0 30 20 ? * MON") {

        Date minDate = new Date().minus(RemoteCounterConstants.RETENTION_IN_DAYS)
        java.text.SimpleDateFormat sdf = new java.text.SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSSXXX");
        sdf.setTimeZone(TimeZone.getTimeZone("CET"));
        String minDateAsStr = sdf.format(minDate);

        log.debug("Deleting items older than " + minDateAsStr)
                    log.warn("current user = ${security.currentUser().getUsername()}")

        ((Searches) searches).aql(
            "items.find({" +
                "\"repo\": \"" + RemoteCounterConstants.LOCAL_COUNTER_REPO + "\"," +
                "\"created\" : {\"\$lt\" :\"" + minDateAsStr + "\"}" +
            "})") {
            AqlResult result ->
                result.each { b ->
                    log.debug(b.toString())

                    RepoPath rp = RepoPathFactory.create(b.repo, b.path)

                    log.debug("Deleting $rp")
                    repositories.delete(rp)
            }
        }
    }
}
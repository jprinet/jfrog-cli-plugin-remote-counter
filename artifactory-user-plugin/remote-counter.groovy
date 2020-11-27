import org.artifactory.request.Request
import org.artifactory.repo.RepoPath
import org.artifactory.repo.RepoPathFactory
import org.artifactory.security.User

/**
 * This plugin creates a file named after the current timestamp each time a remote download happens.
 * The file is stored in remote-counter-local/${username}.
 *
 * Any user should be allowed to deploy to remote-counter-local
 *
 */

download {
    afterRemoteDownload { request, repoPath ->
        try {
            // uncomment if you wish to filter by name the repositories to monitor
            //if (repoKey.contains("docker")) {

            String localCounterRepo = "remote-counter-local"
            log.debug("request = ${request}")
            log.debug("repoPath = ${repoPath}")

            String username = security.currentUser().getUsername()
            log.debug("user = ${username}")

            String timestamp = java.time.LocalDateTime.now()
            log.debug("timestamp = " + timestamp)

            // instantiate counter file
            RepoPath counterFile = RepoPathFactory.create(localCounterRepo, username + "/" + repoPath.getRepoKey() + "/" + timestamp)

            // deploy counter file
            repositories.deploy(counterFile, new java.io.StringBufferInputStream(""))

            //}
        } catch(Exception e){
            log.error("Error in remote-counter", e);
        }
    }

    //TODO archive old entries
}
# mmctl no longer lives in this repository since we have moved to using a monorepo located in https://github.com/mattermost/mattermost.

[Our developer setup instructions](https://developers.mattermost.com/contribute/developer-setup/) have been updated to use the monorepo, and we're going to be continuing to update our developer documentation to reflect the new setup over the next couple weeks.

All changes going forward should be made in the monorepo since we're no longer accepting PRs in this repository, and any existing PRs should be resubmitted over there. We have some notes on how to do that migration [here](https://developers.mattermost.com/contribute/monorepo-migration-notes/), but most of the code that was previously in this repository is now located in [`server/cmd/mmctl`](https://github.com/mattermost/mattermost-server/tree/master/server/cmd/mmctl) in the monorepo. We were unable to maintain Git history with the move, so migrating PRs to the new repo will likely involve a lot of manually copying changes to their new locations.

This repository is being kept open until December 2023 to maintain support for our [extended support releases](https://docs.mattermost.com/upgrade/extended-support-release.html) at which point it will be archived and kept as a historical record.

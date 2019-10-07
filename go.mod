module github.com/mattermost/mmctl

go 1.12

require (
	github.com/magefile/mage v1.8.0
	github.com/mattermost/mattermost-server v0.0.0-20190417144445-84a59ddb3928
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.3
	golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c
	golang.org/x/tools v0.0.0-20191007151716-af92b0ef1f12 // indirect
)

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	// Workaround for https://github.com/golang/go/issues/30831 and fallout.
	github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1
)

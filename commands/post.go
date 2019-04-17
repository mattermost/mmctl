package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var PostCmd = &cobra.Command{
	Use:   "post",
	Short: "Management of posts",
}

var PostCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a post",
	Example: `  post create myteam:mychannel --message "some text for the post"`,
	Args:    cobra.ExactArgs(1),
	RunE:    postCreateCmdF,
}

var PostListCmd = &cobra.Command{
	Use:   "list",
	Short: "List posts for a channel",
	Example: `  post list myteam:mychannel
  post list myteam:mychannel --number 20`,
	Args: cobra.ExactArgs(1),
	RunE: postListCmdF,
}

func init() {
	PostCreateCmd.Flags().StringP("message", "m", "", "Message for the post")
	PostCreateCmd.Flags().StringP("reply-to", "r", "", "Post id to reply to")

	PostListCmd.Flags().IntP("number", "n", 20, "Number of messages to list")

	PostCmd.AddCommand(
		PostCreateCmd,
		PostListCmd,
	)

	RootCmd.AddCommand(PostCmd)
}

func postCreateCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	message, _ := command.Flags().GetString("message")
	if message == "" {
		return errors.New("Message cannot be empty")
	}

	replyTo, _ := command.Flags().GetString("reply-to")
	if replyTo != "" {
		replyToPost, res := c.GetPost(replyTo, "")
		if res.Error != nil {
			return res.Error
		}
		if replyToPost.RootId != "" {
			replyTo = replyToPost.RootId
		}
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   message,
		RootId:    replyTo,
	}

	if _, res := c.CreatePost(post); res.Error != nil {
		return res.Error
	}
	return nil
}

func postListCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	number, _ := command.Flags().GetInt("number")

	postList, res := c.GetPostsForChannel(channel.Id, 0, number, "")
	if res.Error != nil {
		return res.Error
	}

	posts := postList.ToSlice()
	for i := 1; i <= len(posts); i++ {
		post := posts[len(posts)-i]
		var username string
		user, res := c.GetUser(post.UserId, "")
		if res.Error != nil {
			username = post.UserId
		} else {
			username = user.Username
		}

		fmt.Println(fmt.Sprintf("%s [%s] %s", post.Id, username, post.Message))
	}
	return nil
}

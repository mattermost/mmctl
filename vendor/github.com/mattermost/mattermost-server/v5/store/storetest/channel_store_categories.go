// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetest

import (
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

func TestChannelStoreCategories(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("CreateInitialSidebarCategories", func(t *testing.T) { testCreateInitialSidebarCategories(t, ss) })
	t.Run("CreateSidebarCategory", func(t *testing.T) { testCreateSidebarCategory(t, ss) })
	t.Run("GetSidebarCategory", func(t *testing.T) { testGetSidebarCategory(t, ss, s) })
	t.Run("GetSidebarCategories", func(t *testing.T) { testGetSidebarCategories(t, ss) })
	t.Run("UpdateSidebarCategories", func(t *testing.T) { testUpdateSidebarCategories(t, ss) })
	t.Run("ClearSidebarOnTeamLeave", func(t *testing.T) { testClearSidebarOnTeamLeave(t, ss, s) })
	t.Run("DeleteSidebarCategory", func(t *testing.T) { testDeleteSidebarCategory(t, ss, s) })
	t.Run("UpdateSidebarChannelsByPreferences", func(t *testing.T) { testUpdateSidebarChannelsByPreferences(t, ss) })
	t.Run("SidebarCategoryDeadlock", func(t *testing.T) { testSidebarCategoryDeadlock(t, ss) })
}

func testCreateInitialSidebarCategories(t *testing.T, ss store.Store) {
	t.Run("should create initial favorites/channels/DMs categories", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		assert.NoError(t, nErr)
		assert.Len(t, res.Categories, 3)
		assert.Equal(t, model.SidebarCategoryFavorites, res.Categories[0].Type)
		assert.Equal(t, model.SidebarCategoryChannels, res.Categories[1].Type)
		assert.Equal(t, model.SidebarCategoryDirectMessages, res.Categories[2].Type)

		res2, err := ss.Channel().GetSidebarCategories(userId, teamId)
		assert.NoError(t, err)
		assert.Equal(t, res, res2)
	})

	t.Run("should create initial favorites/channels/DMs categories for multiple users", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		userId2 := model.NewId()

		res, nErr = ss.Channel().CreateInitialSidebarCategories(userId2, teamId)
		assert.NoError(t, nErr)
		assert.Len(t, res.Categories, 3)
		assert.Equal(t, model.SidebarCategoryFavorites, res.Categories[0].Type)
		assert.Equal(t, model.SidebarCategoryChannels, res.Categories[1].Type)
		assert.Equal(t, model.SidebarCategoryDirectMessages, res.Categories[2].Type)

		res2, err := ss.Channel().GetSidebarCategories(userId2, teamId)
		assert.NoError(t, err)
		assert.Equal(t, res, res2)
	})

	t.Run("should create initial favorites/channels/DMs categories on different teams", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		teamId2 := model.NewId()

		res, nErr = ss.Channel().CreateInitialSidebarCategories(userId, teamId2)
		assert.NoError(t, nErr)
		assert.Len(t, res.Categories, 3)
		assert.Equal(t, model.SidebarCategoryFavorites, res.Categories[0].Type)
		assert.Equal(t, model.SidebarCategoryChannels, res.Categories[1].Type)
		assert.Equal(t, model.SidebarCategoryDirectMessages, res.Categories[2].Type)

		res2, err := ss.Channel().GetSidebarCategories(userId, teamId2)
		assert.NoError(t, err)
		assert.Equal(t, res, res2)
	})

	t.Run("shouldn't create additional categories when ones already exist", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		initialCategories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Equal(t, res, initialCategories)

		// Calling CreateInitialSidebarCategories a second time shouldn't create any new categories
		res, nErr = ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		assert.NoError(t, nErr)
		assert.NotEmpty(t, res)

		res, err = ss.Channel().GetSidebarCategories(userId, teamId)
		assert.NoError(t, err)
		assert.Equal(t, initialCategories.Categories, res.Categories)
	})

	t.Run("shouldn't create additional categories when ones already exist even when ran simultaneously", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				_, _ = ss.Channel().CreateInitialSidebarCategories(userId, teamId)
			}()
		}

		wg.Wait()

		res, err := ss.Channel().GetSidebarCategories(userId, teamId)
		assert.NoError(t, err)
		assert.Len(t, res.Categories, 3)
	})

	t.Run("should populate the Favorites category with regular channels", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Set up two channels, one favorited and one not
		channel1, nErr := ss.Channel().Save(&model.Channel{
			TeamId: teamId,
			Type:   model.CHANNEL_OPEN,
			Name:   "channel1",
		}, 1000)
		require.NoError(t, nErr)
		_, err := ss.Channel().SaveMember(&model.ChannelMember{
			ChannelId:   channel1.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, err)

		channel2, nErr := ss.Channel().Save(&model.Channel{
			TeamId: teamId,
			Type:   model.CHANNEL_OPEN,
			Name:   "channel2",
		}, 1000)
		require.NoError(t, nErr)
		_, err = ss.Channel().SaveMember(&model.ChannelMember{
			ChannelId:   channel2.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, err)

		nErr = ss.Preference().Save(&model.Preferences{
			{
				UserId:   userId,
				Category: model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL,
				Name:     channel1.Id,
				Value:    "true",
			},
		})
		require.NoError(t, nErr)

		// Create the categories
		categories, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.Len(t, categories.Categories, 3)
		assert.Equal(t, model.SidebarCategoryFavorites, categories.Categories[0].Type)
		assert.Equal(t, []string{channel1.Id}, categories.Categories[0].Channels)
		assert.Equal(t, model.SidebarCategoryChannels, categories.Categories[1].Type)
		assert.Equal(t, []string{channel2.Id}, categories.Categories[1].Channels)

		// Get and check the categories for channels
		categories2, nErr := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.Equal(t, categories, categories2)
	})

	t.Run("should populate the Favorites category in alphabetical order", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Set up two channels
		channel1, nErr := ss.Channel().Save(&model.Channel{
			TeamId:      teamId,
			Type:        model.CHANNEL_OPEN,
			Name:        "channel1",
			DisplayName: "zebra",
		}, 1000)
		require.NoError(t, nErr)
		_, err := ss.Channel().SaveMember(&model.ChannelMember{
			ChannelId:   channel1.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, err)

		channel2, nErr := ss.Channel().Save(&model.Channel{
			TeamId:      teamId,
			Type:        model.CHANNEL_OPEN,
			Name:        "channel2",
			DisplayName: "aardvark",
		}, 1000)
		require.NoError(t, nErr)
		_, err = ss.Channel().SaveMember(&model.ChannelMember{
			ChannelId:   channel2.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, err)

		nErr = ss.Preference().Save(&model.Preferences{
			{
				UserId:   userId,
				Category: model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL,
				Name:     channel1.Id,
				Value:    "true",
			},
			{
				UserId:   userId,
				Category: model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL,
				Name:     channel2.Id,
				Value:    "true",
			},
		})
		require.NoError(t, nErr)

		// Create the categories
		categories, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.Len(t, categories.Categories, 3)
		assert.Equal(t, model.SidebarCategoryFavorites, categories.Categories[0].Type)
		assert.Equal(t, []string{channel2.Id, channel1.Id}, categories.Categories[0].Channels)

		// Get and check the categories for channels
		categories2, nErr := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.Equal(t, categories, categories2)
	})

	t.Run("should populate the Favorites category with DMs and GMs", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		otherUserId1 := model.NewId()
		otherUserId2 := model.NewId()

		// Set up two direct channels, one favorited and one not
		dmChannel1, err := ss.Channel().SaveDirectChannel(
			&model.Channel{
				Name: model.GetDMNameFromIds(userId, otherUserId1),
				Type: model.CHANNEL_DIRECT,
			},
			&model.ChannelMember{
				UserId:      userId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
			&model.ChannelMember{
				UserId:      otherUserId1,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
		)
		require.NoError(t, err)

		dmChannel2, err := ss.Channel().SaveDirectChannel(
			&model.Channel{
				Name: model.GetDMNameFromIds(userId, otherUserId2),
				Type: model.CHANNEL_DIRECT,
			},
			&model.ChannelMember{
				UserId:      userId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
			&model.ChannelMember{
				UserId:      otherUserId2,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
		)
		require.NoError(t, err)

		err = ss.Preference().Save(&model.Preferences{
			{
				UserId:   userId,
				Category: model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL,
				Name:     dmChannel1.Id,
				Value:    "true",
			},
		})
		require.NoError(t, err)

		// Create the categories
		categories, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.Len(t, categories.Categories, 3)
		assert.Equal(t, model.SidebarCategoryFavorites, categories.Categories[0].Type)
		assert.Equal(t, []string{dmChannel1.Id}, categories.Categories[0].Channels)
		assert.Equal(t, model.SidebarCategoryDirectMessages, categories.Categories[2].Type)
		assert.Equal(t, []string{dmChannel2.Id}, categories.Categories[2].Channels)

		// Get and check the categories for channels
		categories2, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Equal(t, categories, categories2)
	})

	t.Run("should not populate the Favorites category with channels from other teams", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()
		teamId2 := model.NewId()

		// Set up a channel on another team and favorite it
		channel1, nErr := ss.Channel().Save(&model.Channel{
			TeamId: teamId2,
			Type:   model.CHANNEL_OPEN,
			Name:   "channel1",
		}, 1000)
		require.NoError(t, nErr)
		_, err := ss.Channel().SaveMember(&model.ChannelMember{
			ChannelId:   channel1.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, err)

		nErr = ss.Preference().Save(&model.Preferences{
			{
				UserId:   userId,
				Category: model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL,
				Name:     channel1.Id,
				Value:    "true",
			},
		})
		require.NoError(t, nErr)

		// Create the categories
		categories, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.Len(t, categories.Categories, 3)
		assert.Equal(t, model.SidebarCategoryFavorites, categories.Categories[0].Type)
		assert.Equal(t, []string{}, categories.Categories[0].Channels)
		assert.Equal(t, model.SidebarCategoryChannels, categories.Categories[1].Type)
		assert.Equal(t, []string{}, categories.Categories[1].Channels)

		// Get and check the categories for channels
		categories2, nErr := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.Equal(t, categories, categories2)
	})
}

func testCreateSidebarCategory(t *testing.T, ss store.Store) {
	t.Run("Creating category without initial categories should fail", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the category
		created, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			SidebarCategory: model.SidebarCategory{
				DisplayName: model.NewId(),
			},
		})
		require.Error(t, err)
		var errNotFound *store.ErrNotFound
		require.ErrorAs(t, err, &errNotFound)
		require.Nil(t, created)
	})

	t.Run("should place the new category second if Favorites comes first", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		// Create the category
		created, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			SidebarCategory: model.SidebarCategory{
				DisplayName: model.NewId(),
			},
		})
		require.NoError(t, err)

		// Confirm that it comes second
		res, err = ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, res.Categories, 4)
		assert.Equal(t, model.SidebarCategoryFavorites, res.Categories[0].Type)
		assert.Equal(t, model.SidebarCategoryCustom, res.Categories[1].Type)
		assert.Equal(t, created.Id, res.Categories[1].Id)
	})

	t.Run("should place the new category first if Favorites is not first", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		// Re-arrange the categories so that Favorites comes last
		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, categories.Categories, 3)
		require.Equal(t, model.SidebarCategoryFavorites, categories.Categories[0].Type)

		err = ss.Channel().UpdateSidebarCategoryOrder(userId, teamId, []string{
			categories.Categories[1].Id,
			categories.Categories[2].Id,
			categories.Categories[0].Id,
		})
		require.NoError(t, err)

		// Create the category
		created, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			SidebarCategory: model.SidebarCategory{
				DisplayName: model.NewId(),
			},
		})
		require.NoError(t, err)

		// Confirm that it comes first
		res, err = ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, res.Categories, 4)
		assert.Equal(t, model.SidebarCategoryCustom, res.Categories[0].Type)
		assert.Equal(t, created.Id, res.Categories[0].Id)
	})

	t.Run("should create the category with its channels", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		// Create some channels
		channel1, err := ss.Channel().Save(&model.Channel{
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
			Name:   model.NewId(),
		}, 100)
		require.NoError(t, err)
		channel2, err := ss.Channel().Save(&model.Channel{
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
			Name:   model.NewId(),
		}, 100)
		require.NoError(t, err)

		// Create the category
		created, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			SidebarCategory: model.SidebarCategory{
				DisplayName: model.NewId(),
			},
			Channels: []string{channel2.Id, channel1.Id},
		})
		require.NoError(t, err)
		assert.Equal(t, []string{channel2.Id, channel1.Id}, created.Channels)

		// Get the channel again to ensure that the SidebarChannels were saved correctly
		res2, err := ss.Channel().GetSidebarCategory(created.Id)
		require.NoError(t, err)
		assert.Equal(t, []string{channel2.Id, channel1.Id}, res2.Channels)
	})

	t.Run("should remove any channels from their previous categories", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, categories.Categories, 3)

		favoritesCategory := categories.Categories[0]
		require.Equal(t, model.SidebarCategoryFavorites, favoritesCategory.Type)
		channelsCategory := categories.Categories[1]
		require.Equal(t, model.SidebarCategoryChannels, channelsCategory.Type)

		// Create some channels
		channel1, nErr := ss.Channel().Save(&model.Channel{
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
			Name:   model.NewId(),
		}, 100)
		require.NoError(t, nErr)
		channel2, nErr := ss.Channel().Save(&model.Channel{
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
			Name:   model.NewId(),
		}, 100)
		require.NoError(t, nErr)

		// Assign them to categories
		favoritesCategory.Channels = []string{channel1.Id}
		channelsCategory.Channels = []string{channel2.Id}
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			favoritesCategory,
			channelsCategory,
		})
		require.NoError(t, err)

		// Create the category
		created, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			SidebarCategory: model.SidebarCategory{
				DisplayName: model.NewId(),
			},
			Channels: []string{channel2.Id, channel1.Id},
		})
		require.NoError(t, err)
		assert.Equal(t, []string{channel2.Id, channel1.Id}, created.Channels)

		// Confirm that the channels were removed from their original categories
		res2, err := ss.Channel().GetSidebarCategory(favoritesCategory.Id)
		require.NoError(t, err)
		assert.Equal(t, []string{}, res2.Channels)

		res2, err = ss.Channel().GetSidebarCategory(channelsCategory.Id)
		require.NoError(t, err)
		assert.Equal(t, []string{}, res2.Channels)
	})
}

func testGetSidebarCategory(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("should return a custom category with its Channels field set", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		channelId1 := model.NewId()
		channelId2 := model.NewId()
		channelId3 := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		// Create a category and assign some channels to it
		created, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			SidebarCategory: model.SidebarCategory{
				UserId:      userId,
				TeamId:      teamId,
				DisplayName: model.NewId(),
			},
			Channels: []string{channelId1, channelId2, channelId3},
		})
		require.NoError(t, err)
		require.NotNil(t, created)

		// Ensure that they're returned in order
		res2, err := ss.Channel().GetSidebarCategory(created.Id)
		assert.NoError(t, err)
		assert.Equal(t, created.Id, res2.Id)
		assert.Equal(t, model.SidebarCategoryCustom, res2.Type)
		assert.Equal(t, created.DisplayName, res2.DisplayName)
		assert.Equal(t, []string{channelId1, channelId2, channelId3}, res2.Channels)
	})

	t.Run("should return any orphaned channels with the Channels category", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the channels category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		channelsCategory := categories.Categories[1]
		require.Equal(t, model.SidebarCategoryChannels, channelsCategory.Type)

		// Join some channels
		channel1, nErr := ss.Channel().Save(&model.Channel{
			Name:        "channel1",
			DisplayName: "DEF",
			TeamId:      teamId,
			Type:        model.CHANNEL_PRIVATE,
		}, 10)
		require.NoError(t, nErr)
		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel1.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)

		channel2, nErr := ss.Channel().Save(&model.Channel{
			Name:        "channel2",
			DisplayName: "ABC",
			TeamId:      teamId,
			Type:        model.CHANNEL_OPEN,
		}, 10)
		require.NoError(t, nErr)
		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel2.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)

		// Confirm that they're not in the Channels category in the DB
		count, countErr := s.GetMaster().SelectInt(`
			SELECT
				COUNT(*)
			FROM
				SidebarChannels
			WHERE
				CategoryId = :CategoryId`, map[string]interface{}{"CategoryId": channelsCategory.Id})
		require.NoError(t, countErr)
		assert.Equal(t, int64(0), count)

		// Ensure that the Channels are returned in alphabetical order
		res2, err := ss.Channel().GetSidebarCategory(channelsCategory.Id)
		assert.NoError(t, err)
		assert.Equal(t, channelsCategory.Id, res2.Id)
		assert.Equal(t, model.SidebarCategoryChannels, channelsCategory.Type)
		assert.Equal(t, []string{channel2.Id, channel1.Id}, res2.Channels)
	})

	t.Run("shouldn't return orphaned channels on another team with the Channels category", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the channels category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Equal(t, model.SidebarCategoryChannels, categories.Categories[1].Type)

		channelsCategory := categories.Categories[1]

		// Join a channel on another team
		channel1, nErr := ss.Channel().Save(&model.Channel{
			Name:   "abc",
			TeamId: model.NewId(),
			Type:   model.CHANNEL_OPEN,
		}, 10)
		require.NoError(t, nErr)
		defer ss.Channel().PermanentDelete(channel1.Id)

		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel1.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)

		// Ensure that no channels are returned
		res2, err := ss.Channel().GetSidebarCategory(channelsCategory.Id)
		assert.NoError(t, err)
		assert.Equal(t, channelsCategory.Id, res2.Id)
		assert.Equal(t, model.SidebarCategoryChannels, channelsCategory.Type)
		assert.Len(t, res2.Channels, 0)
	})

	t.Run("shouldn't return non-orphaned channels with the Channels category", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the channels category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		favoritesCategory := categories.Categories[0]
		require.Equal(t, model.SidebarCategoryFavorites, favoritesCategory.Type)
		channelsCategory := categories.Categories[1]
		require.Equal(t, model.SidebarCategoryChannels, channelsCategory.Type)

		// Join some channels
		channel1, nErr := ss.Channel().Save(&model.Channel{
			Name:        "channel1",
			DisplayName: "DEF",
			TeamId:      teamId,
			Type:        model.CHANNEL_PRIVATE,
		}, 10)
		require.NoError(t, nErr)
		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel1.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)

		channel2, nErr := ss.Channel().Save(&model.Channel{
			Name:        "channel2",
			DisplayName: "ABC",
			TeamId:      teamId,
			Type:        model.CHANNEL_OPEN,
		}, 10)
		require.NoError(t, nErr)
		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel2.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)

		// And assign one to another category
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory.SidebarCategory,
				Channels:        []string{channel2.Id},
			},
		})
		require.NoError(t, err)

		// Ensure that the correct channel is returned in the Channels category
		res2, err := ss.Channel().GetSidebarCategory(channelsCategory.Id)
		assert.NoError(t, err)
		assert.Equal(t, channelsCategory.Id, res2.Id)
		assert.Equal(t, model.SidebarCategoryChannels, channelsCategory.Type)
		assert.Equal(t, []string{channel1.Id}, res2.Channels)
	})

	t.Run("should return any orphaned DM channels with the Direct Messages category", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the DMs category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Equal(t, model.SidebarCategoryDirectMessages, categories.Categories[2].Type)

		dmsCategory := categories.Categories[2]

		// Create a DM
		otherUserId := model.NewId()
		dmChannel, nErr := ss.Channel().SaveDirectChannel(
			&model.Channel{
				Name: model.GetDMNameFromIds(userId, otherUserId),
				Type: model.CHANNEL_DIRECT,
			},
			&model.ChannelMember{
				UserId:      userId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
			&model.ChannelMember{
				UserId:      otherUserId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
		)
		require.NoError(t, nErr)

		// Ensure that the DM is returned
		res2, err := ss.Channel().GetSidebarCategory(dmsCategory.Id)
		assert.NoError(t, err)
		assert.Equal(t, dmsCategory.Id, res2.Id)
		assert.Equal(t, model.SidebarCategoryDirectMessages, res2.Type)
		assert.Equal(t, []string{dmChannel.Id}, res2.Channels)
	})

	t.Run("should return any orphaned GM channels with the Direct Messages category", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the DMs category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Equal(t, model.SidebarCategoryDirectMessages, categories.Categories[2].Type)

		dmsCategory := categories.Categories[2]

		// Create a GM
		gmChannel, nErr := ss.Channel().Save(&model.Channel{
			Name:   "abc",
			TeamId: "",
			Type:   model.CHANNEL_GROUP,
		}, 10)
		require.NoError(t, nErr)
		defer ss.Channel().PermanentDelete(gmChannel.Id)
		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   gmChannel.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)

		// Ensure that the DM is returned
		res2, err := ss.Channel().GetSidebarCategory(dmsCategory.Id)
		assert.NoError(t, err)
		assert.Equal(t, dmsCategory.Id, res2.Id)
		assert.Equal(t, model.SidebarCategoryDirectMessages, res2.Type)
		assert.Equal(t, []string{gmChannel.Id}, res2.Channels)
	})

	t.Run("should return orphaned DM channels in the DMs categorywhich are in a custom category on another team", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the DMs category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Equal(t, model.SidebarCategoryDirectMessages, categories.Categories[2].Type)

		dmsCategory := categories.Categories[2]

		// Create a DM
		otherUserId := model.NewId()
		dmChannel, nErr := ss.Channel().SaveDirectChannel(
			&model.Channel{
				Name: model.GetDMNameFromIds(userId, otherUserId),
				Type: model.CHANNEL_DIRECT,
			},
			&model.ChannelMember{
				UserId:      userId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
			&model.ChannelMember{
				UserId:      otherUserId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
		)
		require.NoError(t, nErr)

		// Create another team and assign the DM to a custom category on that team
		otherTeamId := model.NewId()

		res, nErr = ss.Channel().CreateInitialSidebarCategories(userId, otherTeamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		_, err = ss.Channel().CreateSidebarCategory(userId, otherTeamId, &model.SidebarCategoryWithChannels{
			SidebarCategory: model.SidebarCategory{
				UserId: userId,
				TeamId: teamId,
			},
			Channels: []string{dmChannel.Id},
		})
		require.NoError(t, err)

		// Ensure that the DM is returned with the DMs category on the original team
		res2, err := ss.Channel().GetSidebarCategory(dmsCategory.Id)
		assert.NoError(t, err)
		assert.Equal(t, dmsCategory.Id, res2.Id)
		assert.Equal(t, model.SidebarCategoryDirectMessages, res2.Type)
		assert.Equal(t, []string{dmChannel.Id}, res2.Channels)
	})
}

func testGetSidebarCategories(t *testing.T, ss store.Store) {
	t.Run("should return channels in the same order between different ways of getting categories", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		channelIds := []string{
			model.NewId(),
			model.NewId(),
			model.NewId(),
		}

		newCategory, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			Channels: channelIds,
		})
		require.NoError(t, err)
		require.NotNil(t, newCategory)

		gotCategory, err := ss.Channel().GetSidebarCategory(newCategory.Id)
		require.NoError(t, err)

		res, err = ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, res.Categories, 4)

		require.Equal(t, model.SidebarCategoryCustom, res.Categories[1].Type)

		// This looks unnecessary, but I was getting different results from some of these before
		assert.Equal(t, newCategory.Channels, res.Categories[1].Channels)
		assert.Equal(t, gotCategory.Channels, res.Categories[1].Channels)
		assert.Equal(t, channelIds, res.Categories[1].Channels)
	})
}

func testUpdateSidebarCategories(t *testing.T, ss store.Store) {
	t.Run("ensure the query to update SidebarCategories hasn't been polluted by UpdateSidebarCategoryOrder", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories
		res, err := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.NotEmpty(t, res)

		initialCategories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		favoritesCategory := initialCategories.Categories[0]
		channelsCategory := initialCategories.Categories[1]
		dmsCategory := initialCategories.Categories[2]

		// And then update one of them
		updated, _, err := ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			channelsCategory,
		})
		require.NoError(t, err)
		assert.Equal(t, channelsCategory, updated[0])
		assert.Equal(t, "Channels", updated[0].DisplayName)

		// And then reorder the categories
		err = ss.Channel().UpdateSidebarCategoryOrder(userId, teamId, []string{dmsCategory.Id, favoritesCategory.Id, channelsCategory.Id})
		require.NoError(t, err)

		// Which somehow blanks out stuff because ???
		got, err := ss.Channel().GetSidebarCategory(favoritesCategory.Id)
		require.NoError(t, err)
		assert.Equal(t, "Favorites", got.DisplayName)
	})

	t.Run("categories should be returned in their original order", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories
		res, err := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.NotEmpty(t, res)

		initialCategories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		favoritesCategory := initialCategories.Categories[0]
		channelsCategory := initialCategories.Categories[1]
		dmsCategory := initialCategories.Categories[2]

		// And then update them
		updatedCategories, _, err := ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			favoritesCategory,
			channelsCategory,
			dmsCategory,
		})
		assert.NoError(t, err)
		assert.Equal(t, favoritesCategory.Id, updatedCategories[0].Id)
		assert.Equal(t, channelsCategory.Id, updatedCategories[1].Id)
		assert.Equal(t, dmsCategory.Id, updatedCategories[2].Id)
	})

	t.Run("should silently fail to update read only fields", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		initialCategories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		favoritesCategory := initialCategories.Categories[0]
		channelsCategory := initialCategories.Categories[1]
		dmsCategory := initialCategories.Categories[2]

		customCategory, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{})
		require.NoError(t, err)

		categoriesToUpdate := []*model.SidebarCategoryWithChannels{
			// Try to change the type of Favorites
			{
				SidebarCategory: model.SidebarCategory{
					Id:          favoritesCategory.Id,
					DisplayName: "something else",
				},
				Channels: favoritesCategory.Channels,
			},
			// Try to change the type of Channels
			{
				SidebarCategory: model.SidebarCategory{
					Id:   channelsCategory.Id,
					Type: model.SidebarCategoryDirectMessages,
				},
				Channels: channelsCategory.Channels,
			},
			// Try to change the Channels of DMs
			{
				SidebarCategory: dmsCategory.SidebarCategory,
				Channels:        []string{"fakechannel"},
			},
			// Try to change the UserId/TeamId of a custom category
			{
				SidebarCategory: model.SidebarCategory{
					Id:          customCategory.Id,
					UserId:      model.NewId(),
					TeamId:      model.NewId(),
					Sorting:     customCategory.Sorting,
					DisplayName: customCategory.DisplayName,
				},
				Channels: customCategory.Channels,
			},
		}

		updatedCategories, _, err := ss.Channel().UpdateSidebarCategories(userId, teamId, categoriesToUpdate)
		assert.NoError(t, err)

		assert.NotEqual(t, "Favorites", categoriesToUpdate[0].DisplayName)
		assert.Equal(t, "Favorites", updatedCategories[0].DisplayName)
		assert.NotEqual(t, model.SidebarCategoryChannels, categoriesToUpdate[1].Type)
		assert.Equal(t, model.SidebarCategoryChannels, updatedCategories[1].Type)
		assert.NotEqual(t, []string{}, categoriesToUpdate[2].Channels)
		assert.Equal(t, []string{}, updatedCategories[2].Channels)
		assert.NotEqual(t, userId, categoriesToUpdate[3].UserId)
		assert.Equal(t, userId, updatedCategories[3].UserId)
	})

	t.Run("should add and remove favorites preferences based on the Favorites category", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the favorites category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		favoritesCategory := categories.Categories[0]
		require.Equal(t, model.SidebarCategoryFavorites, favoritesCategory.Type)

		// Join a channel
		channel, nErr := ss.Channel().Save(&model.Channel{
			Name:   "channel",
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
		}, 10)
		require.NoError(t, nErr)
		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)

		// Assign it to favorites
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory.SidebarCategory,
				Channels:        []string{channel.Id},
			},
		})
		assert.NoError(t, err)

		res2, nErr := ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.NoError(t, nErr)
		assert.NotNil(t, res2)
		assert.Equal(t, "true", res2.Value)

		// And then remove it
		channelsCategory := categories.Categories[1]
		require.Equal(t, model.SidebarCategoryChannels, channelsCategory.Type)

		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: channelsCategory.SidebarCategory,
				Channels:        []string{channel.Id},
			},
		})
		assert.NoError(t, err)

		res2, nErr = ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.Error(t, nErr)
		assert.True(t, errors.Is(nErr, sql.ErrNoRows))
		assert.Nil(t, res2)
	})

	t.Run("should add and remove favorites preferences for DMs", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the favorites category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		favoritesCategory := categories.Categories[0]
		require.Equal(t, model.SidebarCategoryFavorites, favoritesCategory.Type)

		// Create a direct channel
		otherUserId := model.NewId()

		dmChannel, nErr := ss.Channel().SaveDirectChannel(
			&model.Channel{
				Name: model.GetDMNameFromIds(userId, otherUserId),
				Type: model.CHANNEL_DIRECT,
			},
			&model.ChannelMember{
				UserId:      userId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
			&model.ChannelMember{
				UserId:      otherUserId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
		)
		assert.NoError(t, nErr)

		// Assign it to favorites
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory.SidebarCategory,
				Channels:        []string{dmChannel.Id},
			},
		})
		assert.NoError(t, err)

		res2, nErr := ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, dmChannel.Id)
		assert.NoError(t, nErr)
		assert.NotNil(t, res2)
		assert.Equal(t, "true", res2.Value)

		// And then remove it
		dmsCategory := categories.Categories[2]
		require.Equal(t, model.SidebarCategoryDirectMessages, dmsCategory.Type)

		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: dmsCategory.SidebarCategory,
				Channels:        []string{dmChannel.Id},
			},
		})
		assert.NoError(t, err)

		res2, nErr = ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, dmChannel.Id)
		assert.Error(t, nErr)
		assert.True(t, errors.Is(nErr, sql.ErrNoRows))
		assert.Nil(t, res2)
	})

	t.Run("should add and remove favorites preferences, even if the channel is already favorited in preferences", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()
		teamId2 := model.NewId()

		// Create the initial categories and find the favorites categories in each team
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		favoritesCategory := categories.Categories[0]
		require.Equal(t, model.SidebarCategoryFavorites, favoritesCategory.Type)

		res, nErr = ss.Channel().CreateInitialSidebarCategories(userId, teamId2)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories2, err := ss.Channel().GetSidebarCategories(userId, teamId2)
		require.NoError(t, err)

		favoritesCategory2 := categories2.Categories[0]
		require.Equal(t, model.SidebarCategoryFavorites, favoritesCategory2.Type)

		// Create a direct channel
		otherUserId := model.NewId()

		dmChannel, nErr := ss.Channel().SaveDirectChannel(
			&model.Channel{
				Name: model.GetDMNameFromIds(userId, otherUserId),
				Type: model.CHANNEL_DIRECT,
			},
			&model.ChannelMember{
				UserId:      userId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
			&model.ChannelMember{
				UserId:      otherUserId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
		)
		assert.NoError(t, nErr)

		// Assign it to favorites on the first team. The favorites preference gets set for all teams.
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory.SidebarCategory,
				Channels:        []string{dmChannel.Id},
			},
		})
		assert.NoError(t, err)

		res2, nErr := ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, dmChannel.Id)
		assert.NoError(t, nErr)
		assert.NotNil(t, res2)
		assert.Equal(t, "true", res2.Value)

		// Assign it to favorites on the second team. The favorites preference is already set.
		updated, _, err := ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory2.SidebarCategory,
				Channels:        []string{dmChannel.Id},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{dmChannel.Id}, updated[0].Channels)

		res2, nErr = ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, dmChannel.Id)
		assert.NoError(t, nErr)
		assert.NotNil(t, res2)
		assert.Equal(t, "true", res2.Value)

		// Remove it from favorites on the first team. This clears the favorites preference for all teams.
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory.SidebarCategory,
				Channels:        []string{},
			},
		})
		assert.NoError(t, err)

		res2, nErr = ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, dmChannel.Id)
		require.Error(t, nErr)
		assert.Nil(t, res2)

		// Remove it from favorites on the second team. The favorites preference was already deleted.
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId2, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory2.SidebarCategory,
				Channels:        []string{},
			},
		})
		assert.NoError(t, err)

		res2, nErr = ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, dmChannel.Id)
		require.Error(t, nErr)
		assert.Nil(t, res2)
	})

	t.Run("should not affect other users' favorites preferences", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create the initial categories and find the favorites category
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		favoritesCategory := categories.Categories[0]
		require.Equal(t, model.SidebarCategoryFavorites, favoritesCategory.Type)
		channelsCategory := categories.Categories[1]
		require.Equal(t, model.SidebarCategoryChannels, channelsCategory.Type)

		// Create the other users' categories
		userId2 := model.NewId()

		res, nErr = ss.Channel().CreateInitialSidebarCategories(userId2, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		categories2, err := ss.Channel().GetSidebarCategories(userId2, teamId)
		require.NoError(t, err)

		favoritesCategory2 := categories2.Categories[0]
		require.Equal(t, model.SidebarCategoryFavorites, favoritesCategory2.Type)
		channelsCategory2 := categories2.Categories[1]
		require.Equal(t, model.SidebarCategoryChannels, channelsCategory2.Type)

		// Have both users join a channel
		channel, nErr := ss.Channel().Save(&model.Channel{
			Name:   "channel",
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
		}, 10)
		require.NoError(t, nErr)
		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)
		_, nErr = ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId2,
			ChannelId:   channel.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, nErr)

		// Have user1 favorite it
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory.SidebarCategory,
				Channels:        []string{channel.Id},
			},
			{
				SidebarCategory: channelsCategory.SidebarCategory,
				Channels:        []string{},
			},
		})
		assert.NoError(t, err)

		res2, nErr := ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.NoError(t, nErr)
		assert.NotNil(t, res2)
		assert.Equal(t, "true", res2.Value)

		res2, nErr = ss.Preference().Get(userId2, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.True(t, errors.Is(nErr, sql.ErrNoRows))
		assert.Nil(t, res2)

		// And user2 favorite it
		_, _, err = ss.Channel().UpdateSidebarCategories(userId2, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: favoritesCategory2.SidebarCategory,
				Channels:        []string{channel.Id},
			},
			{
				SidebarCategory: channelsCategory2.SidebarCategory,
				Channels:        []string{},
			},
		})
		assert.NoError(t, err)

		res2, nErr = ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.NoError(t, nErr)
		assert.NotNil(t, res2)
		assert.Equal(t, "true", res2.Value)

		res2, nErr = ss.Preference().Get(userId2, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.NoError(t, nErr)
		assert.NotNil(t, res2)
		assert.Equal(t, "true", res2.Value)

		// And then user1 unfavorite it
		_, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: channelsCategory.SidebarCategory,
				Channels:        []string{channel.Id},
			},
			{
				SidebarCategory: favoritesCategory.SidebarCategory,
				Channels:        []string{},
			},
		})
		assert.NoError(t, err)

		res2, nErr = ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.True(t, errors.Is(nErr, sql.ErrNoRows))
		assert.Nil(t, res2)

		res2, nErr = ss.Preference().Get(userId2, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.NoError(t, nErr)
		assert.NotNil(t, res2)
		assert.Equal(t, "true", res2.Value)

		// And finally user2 favorite it
		_, _, err = ss.Channel().UpdateSidebarCategories(userId2, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: channelsCategory2.SidebarCategory,
				Channels:        []string{channel.Id},
			},
			{
				SidebarCategory: favoritesCategory2.SidebarCategory,
				Channels:        []string{},
			},
		})
		assert.NoError(t, err)

		res2, nErr = ss.Preference().Get(userId, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.True(t, errors.Is(nErr, sql.ErrNoRows))
		assert.Nil(t, res2)

		res2, nErr = ss.Preference().Get(userId2, model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL, channel.Id)
		assert.True(t, errors.Is(nErr, sql.ErrNoRows))
		assert.Nil(t, res2)
	})

	t.Run("channels removed from Channels or DMs categories should be re-added", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Create some channels
		channel, nErr := ss.Channel().Save(&model.Channel{
			Name:   "channel",
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
		}, 10)
		require.NoError(t, nErr)
		_, err := ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, err)

		otherUserId := model.NewId()
		dmChannel, nErr := ss.Channel().SaveDirectChannel(
			&model.Channel{
				Name: model.GetDMNameFromIds(userId, otherUserId),
				Type: model.CHANNEL_DIRECT,
			},
			&model.ChannelMember{
				UserId:      userId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
			&model.ChannelMember{
				UserId:      otherUserId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
		)
		require.NoError(t, nErr)

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		// And some categories
		initialCategories, nErr := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, nErr)

		channelsCategory := initialCategories.Categories[1]
		dmsCategory := initialCategories.Categories[2]

		require.Equal(t, []string{channel.Id}, channelsCategory.Channels)
		require.Equal(t, []string{dmChannel.Id}, dmsCategory.Channels)

		// Try to save the categories with no channels in them
		categoriesToUpdate := []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: channelsCategory.SidebarCategory,
				Channels:        []string{},
			},
			{
				SidebarCategory: dmsCategory.SidebarCategory,
				Channels:        []string{},
			},
		}

		updatedCategories, _, nErr := ss.Channel().UpdateSidebarCategories(userId, teamId, categoriesToUpdate)
		assert.NoError(t, nErr)

		// The channels should still exist in the category because they would otherwise be orphaned
		assert.Equal(t, []string{channel.Id}, updatedCategories[0].Channels)
		assert.Equal(t, []string{dmChannel.Id}, updatedCategories[1].Channels)
	})

	t.Run("should be able to move DMs into and out of custom categories", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		otherUserId := model.NewId()
		dmChannel, nErr := ss.Channel().SaveDirectChannel(
			&model.Channel{
				Name: model.GetDMNameFromIds(userId, otherUserId),
				Type: model.CHANNEL_DIRECT,
			},
			&model.ChannelMember{
				UserId:      userId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
			&model.ChannelMember{
				UserId:      otherUserId,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
			},
		)
		require.NoError(t, nErr)

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		// The DM should start in the DMs category
		initialCategories, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)

		dmsCategory := initialCategories.Categories[2]
		require.Equal(t, []string{dmChannel.Id}, dmsCategory.Channels)

		// Now move the DM into a custom category
		customCategory, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{})
		require.NoError(t, err)

		categoriesToUpdate := []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: dmsCategory.SidebarCategory,
				Channels:        []string{},
			},
			{
				SidebarCategory: customCategory.SidebarCategory,
				Channels:        []string{dmChannel.Id},
			},
		}

		updatedCategories, _, err := ss.Channel().UpdateSidebarCategories(userId, teamId, categoriesToUpdate)
		assert.NoError(t, err)
		assert.Equal(t, dmsCategory.Id, updatedCategories[0].Id)
		assert.Equal(t, []string{}, updatedCategories[0].Channels)
		assert.Equal(t, customCategory.Id, updatedCategories[1].Id)
		assert.Equal(t, []string{dmChannel.Id}, updatedCategories[1].Channels)

		updatedDmsCategory, err := ss.Channel().GetSidebarCategory(dmsCategory.Id)
		require.NoError(t, err)
		assert.Equal(t, []string{}, updatedDmsCategory.Channels)

		updatedCustomCategory, err := ss.Channel().GetSidebarCategory(customCategory.Id)
		require.NoError(t, err)
		assert.Equal(t, []string{dmChannel.Id}, updatedCustomCategory.Channels)

		// And move it back out of the custom category
		categoriesToUpdate = []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: dmsCategory.SidebarCategory,
				Channels:        []string{dmChannel.Id},
			},
			{
				SidebarCategory: customCategory.SidebarCategory,
				Channels:        []string{},
			},
		}

		updatedCategories, _, err = ss.Channel().UpdateSidebarCategories(userId, teamId, categoriesToUpdate)
		assert.NoError(t, err)
		assert.Equal(t, dmsCategory.Id, updatedCategories[0].Id)
		assert.Equal(t, []string{dmChannel.Id}, updatedCategories[0].Channels)
		assert.Equal(t, customCategory.Id, updatedCategories[1].Id)
		assert.Equal(t, []string{}, updatedCategories[1].Channels)

		updatedDmsCategory, err = ss.Channel().GetSidebarCategory(dmsCategory.Id)
		require.NoError(t, err)
		assert.Equal(t, []string{dmChannel.Id}, updatedDmsCategory.Channels)

		updatedCustomCategory, err = ss.Channel().GetSidebarCategory(customCategory.Id)
		require.NoError(t, err)
		assert.Equal(t, []string{}, updatedCustomCategory.Channels)
	})

	t.Run("should successfully move channels between categories", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Join a channel
		channel, nErr := ss.Channel().Save(&model.Channel{
			Name:   "channel",
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
		}, 10)
		require.NoError(t, nErr)
		_, err := ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, err)

		// And then create the initial categories so that it includes the channel
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		initialCategories, nErr := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, nErr)

		channelsCategory := initialCategories.Categories[1]
		require.Equal(t, []string{channel.Id}, channelsCategory.Channels)

		customCategory, nErr := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{})
		require.NoError(t, nErr)

		// Move the channel one way
		updatedCategories, _, nErr := ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: channelsCategory.SidebarCategory,
				Channels:        []string{},
			},
			{
				SidebarCategory: customCategory.SidebarCategory,
				Channels:        []string{channel.Id},
			},
		})
		assert.NoError(t, nErr)

		assert.Equal(t, []string{}, updatedCategories[0].Channels)
		assert.Equal(t, []string{channel.Id}, updatedCategories[1].Channels)

		// And then the other
		updatedCategories, _, nErr = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: channelsCategory.SidebarCategory,
				Channels:        []string{channel.Id},
			},
			{
				SidebarCategory: customCategory.SidebarCategory,
				Channels:        []string{},
			},
		})
		assert.NoError(t, nErr)
		assert.Equal(t, []string{channel.Id}, updatedCategories[0].Channels)
		assert.Equal(t, []string{}, updatedCategories[1].Channels)
	})

	t.Run("should correctly return the original categories that were modified", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		// Join a channel
		channel, nErr := ss.Channel().Save(&model.Channel{
			Name:   "channel",
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
		}, 10)
		require.NoError(t, nErr)
		_, err := ss.Channel().SaveMember(&model.ChannelMember{
			UserId:      userId,
			ChannelId:   channel.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		require.NoError(t, err)

		// And then create the initial categories so that Channels includes the channel
		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		initialCategories, nErr := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, nErr)

		channelsCategory := initialCategories.Categories[1]
		require.Equal(t, []string{channel.Id}, channelsCategory.Channels)

		customCategory, nErr := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			SidebarCategory: model.SidebarCategory{
				DisplayName: "originalName",
			},
		})
		require.NoError(t, nErr)

		// Rename the custom category
		updatedCategories, originalCategories, nErr := ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: model.SidebarCategory{
					Id:          customCategory.Id,
					DisplayName: "updatedName",
				},
			},
		})
		require.NoError(t, nErr)
		require.Equal(t, len(updatedCategories), len(originalCategories))
		assert.Equal(t, "originalName", originalCategories[0].DisplayName)
		assert.Equal(t, "updatedName", updatedCategories[0].DisplayName)

		// Move a channel
		updatedCategories, originalCategories, nErr = ss.Channel().UpdateSidebarCategories(userId, teamId, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: channelsCategory.SidebarCategory,
				Channels:        []string{},
			},
			{
				SidebarCategory: customCategory.SidebarCategory,
				Channels:        []string{channel.Id},
			},
		})
		require.NoError(t, nErr)
		require.Equal(t, len(updatedCategories), len(originalCategories))
		require.Equal(t, updatedCategories[0].Id, originalCategories[0].Id)
		require.Equal(t, updatedCategories[1].Id, originalCategories[1].Id)

		assert.Equal(t, []string{channel.Id}, originalCategories[0].Channels)
		assert.Equal(t, []string{}, updatedCategories[0].Channels)
		assert.Equal(t, []string{}, originalCategories[1].Channels)
		assert.Equal(t, []string{channel.Id}, updatedCategories[1].Channels)
	})
}

func setupInitialSidebarCategories(t *testing.T, ss store.Store) (string, string) {
	userId := model.NewId()
	teamId := model.NewId()

	res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
	require.NoError(t, nErr)
	require.NotEmpty(t, res)

	res, err := ss.Channel().GetSidebarCategories(userId, teamId)
	require.NoError(t, err)
	require.Len(t, res.Categories, 3)

	return userId, teamId
}

func testClearSidebarOnTeamLeave(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("should delete all sidebar categories and channels on the team", func(t *testing.T) {
		userId, teamId := setupInitialSidebarCategories(t, ss)

		user := &model.User{
			Id: userId,
		}

		// Create some channels and assign them to a custom category
		channel1, nErr := ss.Channel().Save(&model.Channel{
			Name:   model.NewId(),
			TeamId: teamId,
			Type:   model.CHANNEL_OPEN,
		}, 1000)
		require.NoError(t, nErr)

		dmChannel1, nErr := ss.Channel().CreateDirectChannel(user, &model.User{
			Id: model.NewId(),
		})
		require.NoError(t, nErr)

		_, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			Channels: []string{channel1.Id, dmChannel1.Id},
		})
		require.NoError(t, err)

		// Confirm that we start with the right number of categories and SidebarChannels entries
		count, err := s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarCategories WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		require.Equal(t, int64(4), count)

		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarChannels WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		require.Equal(t, int64(2), count)

		// Leave the team
		err = ss.Channel().ClearSidebarOnTeamLeave(userId, teamId)
		assert.NoError(t, err)

		// Confirm that all the categories and SidebarChannel entries have been deleted
		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarCategories WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarChannels WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should not delete sidebar categories and channels on another the team", func(t *testing.T) {
		userId, teamId := setupInitialSidebarCategories(t, ss)

		user := &model.User{
			Id: userId,
		}

		// Create some channels and assign them to a custom category
		channel1, nErr := ss.Channel().Save(&model.Channel{
			Name:   model.NewId(),
			TeamId: teamId,
			Type:   model.CHANNEL_OPEN,
		}, 1000)
		require.NoError(t, nErr)

		dmChannel1, nErr := ss.Channel().CreateDirectChannel(user, &model.User{
			Id: model.NewId(),
		})
		require.NoError(t, nErr)

		_, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			Channels: []string{channel1.Id, dmChannel1.Id},
		})
		require.NoError(t, err)

		// Confirm that we start with the right number of categories and SidebarChannels entries
		count, err := s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarCategories WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		require.Equal(t, int64(4), count)

		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarChannels WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		require.Equal(t, int64(2), count)

		// Leave another team
		err = ss.Channel().ClearSidebarOnTeamLeave(userId, model.NewId())
		assert.NoError(t, err)

		// Confirm that nothing has been deleted
		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarCategories WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		assert.Equal(t, int64(4), count)

		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarChannels WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("MM-30314 should not delete channels on another team under specific circumstances", func(t *testing.T) {
		userId, teamId := setupInitialSidebarCategories(t, ss)

		user := &model.User{
			Id: userId,
		}
		user2 := &model.User{
			Id: model.NewId(),
		}

		// Create a second team and set up the sidebar categories for it
		teamId2 := model.NewId()

		res, err := ss.Channel().CreateInitialSidebarCategories(userId, teamId2)
		require.NoError(t, err)
		require.NotEmpty(t, res)

		res, err = ss.Channel().GetSidebarCategories(userId, teamId2)
		require.NoError(t, err)
		require.Len(t, res.Categories, 3)

		// On the first team, create some channels and assign them to a custom category
		channel1, nErr := ss.Channel().Save(&model.Channel{
			Name:   model.NewId(),
			TeamId: teamId,
			Type:   model.CHANNEL_OPEN,
		}, 1000)
		require.NoError(t, nErr)

		dmChannel1, nErr := ss.Channel().CreateDirectChannel(user, user2)
		require.NoError(t, nErr)

		_, err = ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			Channels: []string{channel1.Id, dmChannel1.Id},
		})
		require.NoError(t, err)

		// Do the same on the second team
		channel2, nErr := ss.Channel().Save(&model.Channel{
			Name:   model.NewId(),
			TeamId: teamId2,
			Type:   model.CHANNEL_OPEN,
		}, 1000)
		require.NoError(t, nErr)

		_, err = ss.Channel().CreateSidebarCategory(userId, teamId2, &model.SidebarCategoryWithChannels{
			Channels: []string{channel2.Id, dmChannel1.Id},
		})
		require.NoError(t, err)

		// Confirm that we start with the right number of categories and SidebarChannels entries
		count, err := s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarCategories WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		require.Equal(t, int64(8), count)

		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarChannels WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		require.Equal(t, int64(4), count)

		// Leave the first team
		err = ss.Channel().ClearSidebarOnTeamLeave(userId, teamId)
		assert.NoError(t, err)

		// Confirm that we have the correct number of categories and SidebarChannels entries left over
		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarCategories WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		assert.Equal(t, int64(4), count)

		count, err = s.GetMaster().SelectInt("SELECT COUNT(*) FROM SidebarChannels WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// Confirm that the categories on the second team are unchanged
		res, err = ss.Channel().GetSidebarCategories(userId, teamId2)
		require.NoError(t, err)
		assert.Len(t, res.Categories, 4)

		assert.Equal(t, model.SidebarCategoryCustom, res.Categories[1].Type)
		assert.Equal(t, []string{channel2.Id, dmChannel1.Id}, res.Categories[1].Channels)
	})
}

func testDeleteSidebarCategory(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("should correctly remove an empty category", func(t *testing.T) {
		userId, teamId := setupInitialSidebarCategories(t, ss)
		defer ss.User().PermanentDelete(userId)

		newCategory, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{})
		require.NoError(t, err)
		require.NotNil(t, newCategory)

		// Ensure that the category was created properly
		res, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, res.Categories, 4)

		// Then delete it and confirm that was done correctly
		err = ss.Channel().DeleteSidebarCategory(newCategory.Id)
		assert.NoError(t, err)

		res, err = ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, res.Categories, 3)
	})

	t.Run("should correctly remove a category and its channels", func(t *testing.T) {
		userId, teamId := setupInitialSidebarCategories(t, ss)
		defer ss.User().PermanentDelete(userId)

		user := &model.User{
			Id: userId,
		}

		// Create some channels
		channel1, nErr := ss.Channel().Save(&model.Channel{
			Name:   model.NewId(),
			TeamId: teamId,
			Type:   model.CHANNEL_OPEN,
		}, 1000)
		require.NoError(t, nErr)
		defer ss.Channel().PermanentDelete(channel1.Id)

		channel2, nErr := ss.Channel().Save(&model.Channel{
			Name:   model.NewId(),
			TeamId: teamId,
			Type:   model.CHANNEL_PRIVATE,
		}, 1000)
		require.NoError(t, nErr)
		defer ss.Channel().PermanentDelete(channel2.Id)

		dmChannel1, nErr := ss.Channel().CreateDirectChannel(user, &model.User{
			Id: model.NewId(),
		})
		require.NoError(t, nErr)
		defer ss.Channel().PermanentDelete(dmChannel1.Id)

		// Assign some of those channels to a custom category
		newCategory, err := ss.Channel().CreateSidebarCategory(userId, teamId, &model.SidebarCategoryWithChannels{
			Channels: []string{channel1.Id, channel2.Id, dmChannel1.Id},
		})
		require.NoError(t, err)
		require.NotNil(t, newCategory)

		// Ensure that the categories are set up correctly
		res, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, res.Categories, 4)

		require.Equal(t, model.SidebarCategoryCustom, res.Categories[1].Type)
		require.Equal(t, []string{channel1.Id, channel2.Id, dmChannel1.Id}, res.Categories[1].Channels)

		// Actually delete the channel
		err = ss.Channel().DeleteSidebarCategory(newCategory.Id)
		assert.NoError(t, err)

		// Confirm that the category was deleted...
		res, err = ss.Channel().GetSidebarCategories(userId, teamId)
		assert.NoError(t, err)
		assert.Len(t, res.Categories, 3)

		// ...and that the corresponding SidebarChannel entries were deleted
		count, countErr := s.GetMaster().SelectInt(`
			SELECT
				COUNT(*)
			FROM
				SidebarChannels
			WHERE
				CategoryId = :CategoryId`, map[string]interface{}{"CategoryId": newCategory.Id})
		require.NoError(t, countErr)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should not allow you to remove non-custom categories", func(t *testing.T) {
		userId, teamId := setupInitialSidebarCategories(t, ss)
		defer ss.User().PermanentDelete(userId)
		res, err := ss.Channel().GetSidebarCategories(userId, teamId)
		require.NoError(t, err)
		require.Len(t, res.Categories, 3)
		require.Equal(t, model.SidebarCategoryFavorites, res.Categories[0].Type)
		require.Equal(t, model.SidebarCategoryChannels, res.Categories[1].Type)
		require.Equal(t, model.SidebarCategoryDirectMessages, res.Categories[2].Type)

		err = ss.Channel().DeleteSidebarCategory(res.Categories[0].Id)
		assert.Error(t, err)

		err = ss.Channel().DeleteSidebarCategory(res.Categories[1].Id)
		assert.Error(t, err)

		err = ss.Channel().DeleteSidebarCategory(res.Categories[2].Id)
		assert.Error(t, err)
	})
}

func testUpdateSidebarChannelsByPreferences(t *testing.T, ss store.Store) {
	t.Run("Should be able to update sidebar channels", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		require.NoError(t, nErr)
		require.NotEmpty(t, res)

		channel, nErr := ss.Channel().Save(&model.Channel{
			Name:   "channel",
			Type:   model.CHANNEL_OPEN,
			TeamId: teamId,
		}, 10)
		require.NoError(t, nErr)

		err := ss.Channel().UpdateSidebarChannelsByPreferences(&model.Preferences{
			model.Preference{
				Name:     channel.Id,
				Category: model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL,
				Value:    "true",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("Should not panic if channel is not found", func(t *testing.T) {
		userId := model.NewId()
		teamId := model.NewId()

		res, nErr := ss.Channel().CreateInitialSidebarCategories(userId, teamId)
		assert.NoError(t, nErr)
		require.NotEmpty(t, res)

		require.NotPanics(t, func() {
			_ = ss.Channel().UpdateSidebarChannelsByPreferences(&model.Preferences{
				model.Preference{
					Name:     "fakeid",
					Category: model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL,
					Value:    "true",
				},
			})
		})
	})
}

// testSidebarCategoryDeadlock tries to delete and update a category at the same time
// in the hope of triggering a deadlock. This is a best-effort test case, and is not guaranteed
// to catch a bug.
func testSidebarCategoryDeadlock(t *testing.T, ss store.Store) {
	userID := model.NewId()
	teamID := model.NewId()

	// Join a channel
	channel, err := ss.Channel().Save(&model.Channel{
		Name:   "channel",
		Type:   model.CHANNEL_OPEN,
		TeamId: teamID,
	}, 10)
	require.NoError(t, err)
	_, err = ss.Channel().SaveMember(&model.ChannelMember{
		UserId:      userID,
		ChannelId:   channel.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})
	require.NoError(t, err)

	// And then create the initial categories so that it includes the channel
	res, err := ss.Channel().CreateInitialSidebarCategories(userID, teamID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	initialCategories, err := ss.Channel().GetSidebarCategories(userID, teamID)
	require.NoError(t, err)

	channelsCategory := initialCategories.Categories[1]
	require.Equal(t, []string{channel.Id}, channelsCategory.Channels)

	customCategory, err := ss.Channel().CreateSidebarCategory(userID, teamID, &model.SidebarCategoryWithChannels{})
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _, err := ss.Channel().UpdateSidebarCategories(userID, teamID, []*model.SidebarCategoryWithChannels{
			{
				SidebarCategory: channelsCategory.SidebarCategory,
				Channels:        []string{},
			},
			{
				SidebarCategory: customCategory.SidebarCategory,
				Channels:        []string{channel.Id},
			},
		})
		if err != nil {
			var nfErr *store.ErrNotFound
			require.True(t, errors.As(err, &nfErr))
		}
	}()

	go func() {
		defer wg.Done()
		err := ss.Channel().DeleteSidebarCategory(customCategory.Id)
		require.NoError(t, err)
	}()

	wg.Wait()
}

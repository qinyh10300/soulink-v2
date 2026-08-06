package demoseed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"soulink-backend/internal/store"
)

type featureSeedResult struct {
	PostsCreated  int
	GroupsCreated int
	EventsCreated int
}

type demoCommentSeed struct {
	AuthorEmail   string
	Content       string
	ParentContent string
}

type demoPostSeed struct {
	AuthorEmail string
	Topic       string
	Content     string
	Comments    []demoCommentSeed
}

var demoPosts = []demoPostSeed{
	{
		AuthorEmail: "chenyu@mails.tsinghua.edu.cn",
		Topic:       "campus",
		Content:     "今天傍晚路过大礼堂，风很舒服。有人也想下课后在校内随便走四十分钟吗？不赶路，只聊最近真正让自己开心的事。",
		Comments: []demoCommentSeed{
			{AuthorEmail: DemoEmail, Content: "可以，我大概 19:30 从紫荆出发。"},
			{AuthorEmail: "xuqinghe@mails.tsinghua.edu.cn", Content: "这个提议很喜欢，艺术博物馆那边晚上也很安静。"},
			{AuthorEmail: "chenyu@mails.tsinghua.edu.cn", Content: "那就先定 19:30，大礼堂西侧见。", ParentContent: "可以，我大概 19:30 从紫荆出发。"},
		},
	},
	{
		AuthorEmail: "zhouyan@mails.tsinghua.edu.cn",
		Topic:       "partner",
		Content:     "想找两位愿意连续四周做产品的同学：每周一次用户访谈、一次小版本发布。方向是 AI 帮助科研团队协作，缺工程和设计伙伴。",
		Comments: []demoCommentSeed{
			{AuthorEmail: "chenyu@mails.tsinghua.edu.cn", Content: "端侧和后端都做过，可以先看看你们现有访谈结论。"},
			{AuthorEmail: DemoEmail, Content: "四周目标很清楚，建议第一次先对齐成功指标和退出条件。"},
		},
	},
	{
		AuthorEmail: "xuqinghe@mails.tsinghua.edu.cn",
		Topic:       "study",
		Content:     "李文正馆晚间自习搭子，周一到周四任选两晚。可以各自安静做事，结束前十分钟交换今天完成了什么。",
		Comments: []demoCommentSeed{
			{AuthorEmail: "shenmo@mails.tsinghua.edu.cn", Content: "周二和周四都可以，八点后到。"},
		},
	},
}

func ensureDemoProductFeatures(
	ctx context.Context,
	dataStore *store.Store,
	users map[string]store.User,
	adminEmail string,
	now time.Time,
) (featureSeedResult, error) {
	result := featureSeedResult{}
	for index, postSeed := range demoPosts {
		created, err := ensureDemoPost(ctx, dataStore, users, postSeed, now.Add(-time.Duration(index+1)*time.Hour))
		if err != nil {
			return featureSeedResult{}, err
		}
		if created {
			result.PostsCreated++
		}
	}
	if err := ensureDemoPlazaInvite(ctx, dataStore, users, demoPosts[0], now); err != nil {
		return featureSeedResult{}, err
	}

	owner := users[DemoEmail]
	if adminEmail != "" {
		owner = users[adminEmail]
	}
	primary := users[DemoEmail]
	chenyu := users["chenyu@mails.tsinghua.edu.cn"]
	zhouyan := users["zhouyan@mails.tsinghua.edu.cn"]
	peers := []store.User{primary, chenyu, zhouyan}
	for _, peer := range peers {
		if peer.ID == owner.ID {
			continue
		}
		if _, err := dataStore.EnsureAcceptedFriendshipBetween(ctx, nil, owner.ID, peer.ID, now); err != nil {
			return featureSeedResult{}, fmt.Errorf("ensure demo friendship: %w", err)
		}
	}

	directPeer := primary
	if directPeer.ID == owner.ID {
		directPeer = chenyu
	}
	conversation, err := dataStore.CreateConversation(ctx, nil, owner.ID, directPeer.ID, now.Add(-45*time.Minute))
	if err != nil {
		return featureSeedResult{}, fmt.Errorf("ensure demo conversation: %w", err)
	}
	if err := ensureDirectMessage(ctx, dataStore, conversation.ID, directPeer.ID, "你好，我把新的树洞、群聊和时间表都走了一遍。我们可以在这里继续验收。", now.Add(-35*time.Minute)); err != nil {
		return featureSeedResult{}, err
	}
	if err := ensureDirectMessage(ctx, dataStore, conversation.ID, owner.ID, "收到，我会重点看匿名边界、群权限和邀约流程。", now.Add(-30*time.Minute)); err != nil {
		return featureSeedResult{}, err
	}

	groupCreated, err := ensureDemoGroup(ctx, dataStore, owner, []store.User{directPeer, zhouyan}, now)
	if err != nil {
		return featureSeedResult{}, err
	}
	if groupCreated {
		result.GroupsCreated++
	}

	eventsCreated, err := ensureDemoCalendar(ctx, dataStore, owner, directPeer, chenyu, now)
	if err != nil {
		return featureSeedResult{}, err
	}
	result.EventsCreated = eventsCreated
	return result, nil
}

func ensureDemoPlazaInvite(
	ctx context.Context,
	dataStore *store.Store,
	users map[string]store.User,
	postSeed demoPostSeed,
	now time.Time,
) error {
	owner := users[postSeed.AuthorEmail]
	postID, err := dataStore.FindPlaygroundPostID(ctx, nil, owner.ID, postSeed.Content)
	if err != nil {
		return fmt.Errorf("find demo plaza group post: %w", err)
	}
	if _, err := dataStore.GetPlazaGroupInviteByPost(ctx, nil, postID); err == nil {
		return nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("find demo plaza group invite: %w", err)
	}
	_, err = dataStore.CreatePlazaGroupInvite(ctx, nil, store.CreatePlazaGroupInviteParams{
		PostID: postID, OwnerID: owner.ID, InviteType: "event",
		Title:         "大礼堂晚风散步小组",
		Summary:       "从树洞讨论整理出的校内散步活动。先在匿名区确认时间，成组后再揭示身份并进入临时群。",
		ExtractedTime: "今天 19:30", ExtractedLocation: "大礼堂西侧",
		MinMembers: 3, MaxMembers: 5, RevealStrategy: "on_form",
		Deadline: now.Add(7 * 24 * time.Hour), GroupExpiresAt: now.Add(14 * 24 * time.Hour),
		InferenceSource: "deterministic", InferenceConfidence: 0.78,
		PromptVersion: "plaza-group-draft-v1", Now: now,
	})
	if err != nil {
		return fmt.Errorf("create demo plaza group invite: %w", err)
	}
	return nil
}

func ensureDemoPost(
	ctx context.Context,
	dataStore *store.Store,
	users map[string]store.User,
	seed demoPostSeed,
	createdAt time.Time,
) (bool, error) {
	author := users[seed.AuthorEmail]
	postID, err := dataStore.FindPlaygroundPostID(ctx, nil, author.ID, seed.Content)
	created := false
	if errors.Is(err, store.ErrNotFound) {
		postID, err = dataStore.CreatePlaygroundPost(ctx, nil, store.CreatePlaygroundPostParams{
			AuthorID:  author.ID,
			Topic:     seed.Topic,
			Content:   seed.Content,
			ImageURLs: []string{},
			Now:       createdAt,
		})
		created = err == nil
	}
	if err != nil {
		return false, fmt.Errorf("ensure demo playground post: %w", err)
	}

	existing, err := dataStore.ListPlaygroundComments(ctx, nil, author.ID, postID)
	if err != nil {
		return false, fmt.Errorf("list demo playground comments: %w", err)
	}
	commentIDs := make(map[string]int64, len(existing)+len(seed.Comments))
	for _, comment := range existing {
		commentIDs[comment.Content] = comment.ID
	}
	for index, commentSeed := range seed.Comments {
		if _, ok := commentIDs[commentSeed.Content]; ok {
			continue
		}
		commentAuthor := users[commentSeed.AuthorEmail]
		params := store.CreatePlaygroundCommentParams{
			PostID:   postID,
			AuthorID: commentAuthor.ID,
			Content:  commentSeed.Content,
			Now:      createdAt.Add(time.Duration(index+1) * 5 * time.Minute),
		}
		if commentSeed.ParentContent != "" {
			parentID, ok := commentIDs[commentSeed.ParentContent]
			if !ok {
				return false, fmt.Errorf("demo comment parent not found: %s", commentSeed.ParentContent)
			}
			params.ParentCommentID = sqlNullInt64(parentID)
		}
		commentID, err := dataStore.CreatePlaygroundComment(ctx, nil, params)
		if err != nil {
			return false, fmt.Errorf("create demo playground comment: %w", err)
		}
		if err := dataStore.AddPlaygroundPostCommentCount(ctx, nil, postID, 1); err != nil {
			return false, fmt.Errorf("increment demo post comment count: %w", err)
		}
		if params.ParentCommentID.Valid {
			if err := dataStore.AddPlaygroundCommentReplyCount(ctx, nil, params.ParentCommentID.Int64, 1); err != nil {
				return false, fmt.Errorf("increment demo reply count: %w", err)
			}
		}
		commentIDs[commentSeed.Content] = commentID
	}
	return created, nil
}

func ensureDirectMessage(ctx context.Context, dataStore *store.Store, conversationID, senderID int64, content string, createdAt time.Time) error {
	messages, err := dataStore.ListMessages(ctx, nil, conversationID, 0, 100)
	if err != nil {
		return fmt.Errorf("list demo direct messages: %w", err)
	}
	for _, message := range messages {
		if message.Content == content {
			return nil
		}
	}
	if _, err := dataStore.CreateMessage(ctx, nil, conversationID, senderID, content, "text", createdAt); err != nil {
		return fmt.Errorf("create demo direct message: %w", err)
	}
	return nil
}

func ensureDemoGroup(ctx context.Context, dataStore *store.Store, owner store.User, members []store.User, now time.Time) (bool, error) {
	const groupName = "SouLink 产品共创"
	group, err := dataStore.FindChatGroup(ctx, nil, owner.ID, groupName)
	created := false
	if errors.Is(err, store.ErrNotFound) {
		group, err = dataStore.CreateChatGroup(ctx, nil, owner.ID, groupName, now.Add(-2*time.Hour))
		created = err == nil
	}
	if err != nil {
		return false, fmt.Errorf("ensure demo group: %w", err)
	}
	if err := dataStore.UpsertChatGroupMember(ctx, nil, group.ID, owner.ID, "owner", now.Add(-2*time.Hour)); err != nil {
		return false, fmt.Errorf("ensure demo group owner: %w", err)
	}
	for _, member := range members {
		if member.ID == owner.ID {
			continue
		}
		if err := dataStore.UpsertChatGroupMember(ctx, nil, group.ID, member.ID, "member", now.Add(-2*time.Hour)); err != nil {
			return false, fmt.Errorf("ensure demo group member: %w", err)
		}
	}
	messageSeeds := []struct {
		senderID int64
		content  string
	}{
		{members[0].ID, "树洞信息密度已经很接近真实产品了，我来补几条回复做压力测试。"},
		{owner.ID, "好，群聊这边也试一下实时消息和成员权限。"},
		{members[1].ID, "我会再建一个周末讨论日程，看看邀请能不能正常收到。"},
	}
	existing, err := dataStore.ListChatGroupMessages(ctx, nil, group.ID, 0, 100)
	if err != nil {
		return false, fmt.Errorf("list demo group messages: %w", err)
	}
	existingContent := make(map[string]struct{}, len(existing))
	for _, message := range existing {
		existingContent[message.Content] = struct{}{}
	}
	for index, messageSeed := range messageSeeds {
		if _, ok := existingContent[messageSeed.content]; ok {
			continue
		}
		if _, err := dataStore.CreateChatGroupMessage(
			ctx, nil, group.ID, messageSeed.senderID, messageSeed.content, "text",
			now.Add(time.Duration(index-3)*10*time.Minute),
		); err != nil {
			return false, fmt.Errorf("create demo group message: %w", err)
		}
	}
	return created, nil
}

func ensureDemoCalendar(ctx context.Context, dataStore *store.Store, owner, directPeer, chenyu store.User, now time.Time) (int, error) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("CST", 8*60*60)
	}
	localNow := now.In(location)
	at := func(days, hour, minute int) time.Time {
		return time.Date(localNow.Year(), localNow.Month(), localNow.Day()+days, hour, minute, 0, 0, location).UTC()
	}
	seeds := []struct {
		organizer      store.User
		title          string
		eventType      string
		description    string
		place          string
		startsAt       time.Time
		endsAt         time.Time
		visibility     string
		attendee       *store.User
		attendeeStatus string
	}{
		{owner, "SouLink 体验走查", "project", "一起验收树洞、聊天和邀约流程。", "李兆基科技大楼", at(1, 19, 0), at(1, 20, 30), "public", &directPeer, "accepted"},
		{directPeer, "周末逛艺术博物馆", "date", "从展览开始，时间合适再一起吃饭。", "清华大学艺术博物馆", at(3, 15, 0), at(3, 17, 0), "public", &owner, "pending"},
		{owner, "产品开发深度工作", "project", "", "", at(2, 9, 30), at(2, 12, 0), "private", nil, ""},
		{chenyu, "端侧 AI 午间交流", "meal", "分享最近的端侧模型和产品观察。", "紫荆园", at(2, 12, 0), at(2, 13, 0), "public", nil, ""},
	}
	createdCount := 0
	for _, seed := range seeds {
		event, err := dataStore.FindCalendarEvent(ctx, nil, seed.organizer.ID, seed.title)
		if errors.Is(err, store.ErrNotFound) {
			event, err = dataStore.CreateCalendarEvent(ctx, nil, store.CreateCalendarEventParams{
				OrganizerID: seed.organizer.ID,
				EventType:   seed.eventType,
				Title:       seed.title,
				Description: seed.description,
				Location:    seed.place,
				StartsAt:    seed.startsAt,
				EndsAt:      seed.endsAt,
				Visibility:  seed.visibility,
				Now:         now,
			})
			if err == nil {
				createdCount++
			}
		} else if err == nil {
			event, err = dataStore.UpdateCalendarEvent(ctx, nil, seed.organizer.ID, store.UpdateCalendarEventParams{
				EventID:     event.ID,
				EventType:   seed.eventType,
				Title:       seed.title,
				Description: seed.description,
				Location:    seed.place,
				StartsAt:    seed.startsAt,
				EndsAt:      seed.endsAt,
				Visibility:  seed.visibility,
				Now:         now,
			})
		}
		if err != nil {
			return 0, fmt.Errorf("ensure demo calendar event: %w", err)
		}
		if seed.attendee != nil && seed.attendee.ID != seed.organizer.ID {
			if err := dataStore.AddCalendarAttendee(ctx, nil, event.ID, seed.attendee.ID, now); err != nil {
				return 0, fmt.Errorf("ensure demo calendar attendee: %w", err)
			}
			if seed.attendeeStatus != "pending" {
				if err := dataStore.UpdateCalendarAttendeeStatus(ctx, nil, event.ID, seed.attendee.ID, seed.attendeeStatus, now); err != nil {
					return 0, fmt.Errorf("ensure demo calendar rsvp: %w", err)
				}
			}
		}
	}
	return createdCount, nil
}

func sqlNullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: true}
}

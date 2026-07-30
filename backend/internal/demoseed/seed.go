package demoseed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"soulink-backend/internal/store"
)

const DemoEmail = "demo@mails.tsinghua.edu.cn"

type Config struct {
	Password   string
	BCryptCost int
}

type Result struct {
	UsersCreated   int
	IntentsCreated int
	RoomsCreated   int
}

type userSeed struct {
	Email              string
	Nickname           string
	Department         string
	Grade              string
	ContactValue       string
	Bio                string
	Skills             []string
	Interests          []string
	Availability       string
	PreferredScenes    []string
	CollaborationStyle string
	Boundaries         string
	AIPortrait         string
	RomanticOptIn      bool
}

type intentSeed struct {
	Key             string
	AuthorEmail     string
	Category        string
	Title           string
	Description     string
	Goal            string
	TimeWindow      string
	Location        string
	Visibility      string
	Constraints     []string
	DealBreakers    []string
	DesiredNextStep string
	ProjectMemo     store.ProjectMemo
	PublishedAgo    time.Duration
}

var demoUsers = []userSeed{
	{
		Email:              DemoEmail,
		Nickname:           "林知行",
		Department:         "计算机系",
		Grade:              "博士一年级",
		ContactValue:       "soulink_demo",
		Bio:                "在做 AI 社交产品，喜欢从真实需求出发快速做原型，也愿意在校园里认识不同方向的人。",
		Skills:             []string{"AI 产品", "前端", "用户研究", "创业"},
		Interests:          []string{"AI 工具", "跑步", "产品设计", "科幻"},
		Availability:       "工作日晚间，周末下午",
		PreferredScenes:    []string{"紫荆园", "李文正馆", "东操"},
		CollaborationStyle: "先把问题和验证目标说清楚，再用一周做出可以被真实用户使用的版本。",
		Boundaries:         "不希望未经双方同意交换联系方式；认真关系匹配需要慢一点。",
		AIPortrait:         "偏行动导向，也重视把不确定性说清楚。适合从一个具体问题或共同项目开始认识。",
		RomanticOptIn:      true,
	},
	{
		Email:              "chenyu@mails.tsinghua.edu.cn",
		Nickname:           "陈屿",
		Department:         "电子系",
		Grade:              "硕士二年级",
		ContactValue:       "chenyu_demo",
		Bio:                "做端侧智能和软硬件协同，吃饭时很乐意聊产品，也愿意聊技术之外的生活。",
		Skills:             []string{"嵌入式", "端侧 AI", "后端", "硬件"},
		Interests:          []string{"AI 产品", "摄影", "咖啡"},
		Availability:       "工作日 18:30 以后",
		PreferredScenes:    []string{"紫荆园", "桃李园"},
		CollaborationStyle: "喜欢先做小规模验证，明确接口后各自推进。",
		Boundaries:         "不接受临时推销或未经确认的线下邀约。",
		AIPortrait:         "表达直接、节奏稳定，适合以技术问题或一顿饭作为低压力的第一次交流。",
	},
	{
		Email:              "zhouyan@mails.tsinghua.edu.cn",
		Nickname:           "周言",
		Department:         "经管学院",
		Grade:              "本科四年级",
		ContactValue:       "zhouyan_demo",
		Bio:                "正在验证一款面向科研团队的 AI 协作工具，希望认识长期主义、能互相补位的伙伴。",
		Skills:             []string{"产品", "增长", "商业分析", "用户访谈"},
		Interests:          []string{"创业", "AI 工具", "羽毛球"},
		Availability:       "每周可稳定投入 12 小时",
		PreferredScenes:    []string{"双清大厦", "清芬园"},
		CollaborationStyle: "重视用户证据、周目标和坦诚复盘。",
		Boundaries:         "合作前需要说清投入、权益和退出方式。",
		AIPortrait:         "目标感强但愿意听取反证，适合和工程执行力强、沟通透明的人协作。",
	},
	{
		Email:              "xuqinghe@mails.tsinghua.edu.cn",
		Nickname:           "许清禾",
		Department:         "建筑学院",
		Grade:              "硕士一年级",
		ContactValue:       "xuqinghe_demo",
		Bio:                "白天做空间与交互研究，晚上常在图书馆学习。偏安静，但熟悉以后话很多。",
		Skills:             []string{"空间设计", "交互设计", "研究"},
		Interests:          []string{"电影", "展览", "城市漫步"},
		Availability:       "周一至周四晚间",
		PreferredScenes:    []string{"李文正馆", "艺术博物馆"},
		CollaborationStyle: "希望先各自专注，再留十分钟同步进度。",
		Boundaries:         "第一次见面偏好公共场所，也不希望被催促回复。",
		AIPortrait:         "观察细致、沟通低压力，适合安静共处或围绕一个具体主题慢慢展开交流。",
		RomanticOptIn:      true,
	},
	{
		Email:              "shenmo@mails.tsinghua.edu.cn",
		Nickname:           "沈墨",
		Department:         "交叉信息院",
		Grade:              "本科三年级",
		ContactValue:       "shenmo_demo",
		Bio:                "白天写模型和系统，晚上用跑步清空脑子。希望找到能稳定互相提醒的运动搭子。",
		Skills:             []string{"机器学习", "系统", "Python"},
		Interests:          []string{"跑步", "开源", "桌游"},
		Availability:       "周二、周四 20:00 后",
		PreferredScenes:    []string{"东操", "紫操"},
		CollaborationStyle: "对目标和时间守约，沟通尽量简洁。",
		Boundaries:         "运动节奏以安全和双方状态为先。",
		AIPortrait:         "自驱、守时，适合有固定节奏的共同活动或边做边聊的技术协作。",
	},
}

var demoIntents = []intentSeed{
	{
		Key:             "primary-project",
		AuthorEmail:     DemoEmail,
		Category:        "cofounder",
		Title:           "为 AI 社交产品找一位前端重的 CTO",
		Description:     "已经完成第一版验证，下一步想把意图理解、推荐解释和校园见面闭环做扎实。希望找到工程能力强、愿意每周稳定投入的人。",
		Goal:            "找到可以共同负责技术方向和产品落地的长期伙伴",
		TimeWindow:      "未来两周内先聊一次，每周约 10 小时",
		Location:        "清华校内",
		Visibility:      "candidates",
		Constraints:     []string{"React 或跨端经验", "能独立完成工程交付", "愿意参与用户访谈"},
		DealBreakers:    []string{"只愿意挂名", "无法稳定投入"},
		DesiredNextStep: "先用 45 分钟对齐问题、现状和未来四周目标",
		ProjectMemo: store.ProjectMemo{
			Building:         "面向清华校内真实社交需求的 AI 意图匹配平台",
			Stage:            "已有 V1 用户与可运行 V2 MVP",
			WeeklyTime:       "每周约 10 小时",
			Compensation:     "先明确长期投入与职责，再讨论联合创始人权益",
			RequiredSkills:   []string{"React", "工程架构", "AI 应用"},
			NiceToHaveSkills: []string{"Go", "推荐系统", "校园产品经验"},
			WorkingStyle:     "周目标、快速上线、每周至少一次用户访谈",
			RiskAppetite:     "接受早期不确定性，但不接受长期没有验证",
		},
		PublishedAgo: 30 * time.Minute,
	},
	{
		Key:             "meal",
		AuthorEmail:     "chenyu@mails.tsinghua.edu.cn",
		Category:        "meal",
		Title:           "今晚在紫荆找一个聊 AI 产品的饭搭子",
		Description:     "最近在做端侧 AI，也很好奇社交产品怎么避免只优化停留时长。想找个人边吃饭边交换真实观察。",
		Goal:            "轻松吃顿饭，认识一个对 AI 产品有真实经验的人",
		TimeWindow:      "今晚 19:00 左右",
		Location:        "紫荆园",
		Visibility:      "feed",
		Constraints:     []string{"时间基本合适", "愿意聊具体经历"},
		DesiredNextStep: "先确认时间，再在紫荆园一层见面",
		PublishedAgo:    12 * time.Minute,
	},
	{
		Key:             "cofounder",
		AuthorEmail:     "zhouyan@mails.tsinghua.edu.cn",
		Category:        "project",
		Title:           "为科研协作 AI 工具找全栈工程伙伴",
		Description:     "已经访谈了十几位实验室同学，想用四周做出可进入真实工作流的版本。需要能一起判断取舍、快速实现的人。",
		Goal:            "完成首个可在实验室内试用的产品版本",
		TimeWindow:      "未来一个月，每周 8 到 12 小时",
		Location:        "双清大厦或线上协作",
		Visibility:      "feed",
		Constraints:     []string{"全栈开发", "AI 应用经验", "稳定投入"},
		DealBreakers:    []string{"目标不透明", "长期失联"},
		DesiredNextStep: "看一遍项目 memo，再做一次技术和协作方式对齐",
		ProjectMemo: store.ProjectMemo{
			Building:         "进入科研团队真实工作流的 AI 协作工具",
			Stage:            "完成十余次访谈，准备开发可用原型",
			WeeklyTime:       "每周 8 到 12 小时",
			Compensation:     "根据持续投入讨论联合创始人权益",
			RequiredSkills:   []string{"全栈开发", "AI 应用", "数据工程"},
			NiceToHaveSkills: []string{"科研工作流", "React", "Go"},
			WorkingStyle:     "以用户证据排优先级，每周交付和复盘",
			RiskAppetite:     "愿意快速试错，但重要假设必须有验证计划",
		},
		PublishedAgo: 48 * time.Minute,
	},
	{
		Key:             "study",
		AuthorEmail:     "xuqinghe@mails.tsinghua.edu.cn",
		Category:        "study",
		Title:           "考试周找一个安静高效的图书馆学习搭子",
		Description:     "想在李文正馆各自专注学习，每 50 分钟短暂休息一次，不需要一直聊天，但希望彼此守时。",
		Goal:            "建立三个晚上的稳定学习节奏",
		TimeWindow:      "本周一到周四 19:30 到 22:00",
		Location:        "李文正馆",
		Visibility:      "feed",
		Constraints:     []string{"偏安静", "守时", "接受番茄钟"},
		DesiredNextStep: "先约一个晚上试一次",
		PublishedAgo:    75 * time.Minute,
	},
	{
		Key:             "sport",
		AuthorEmail:     "shenmo@mails.tsinghua.edu.cn",
		Category:        "sport",
		Title:           "每周二四晚在东操跑五公里",
		Description:     "配速大约六分半，不追求竞速，主要想稳定出门。可以边热身边聊，跑步时各自按状态调整。",
		Goal:            "找到能稳定互相提醒的跑步搭子",
		TimeWindow:      "每周二、周四 20:30",
		Location:        "东操",
		Visibility:      "feed",
		Constraints:     []string{"五公里左右", "以安全为先"},
		DesiredNextStep: "先约下一次跑步，在东操入口碰面",
		PublishedAgo:    2 * time.Hour,
	},
	{
		Key:             "romantic",
		AuthorEmail:     "xuqinghe@mails.tsinghua.edu.cn",
		Category:        "romantic",
		Title:           "想认真认识一个愿意慢慢了解彼此的人",
		Description:     "比起快速判断，更希望从散步、展览或一杯咖啡开始，看看是否能自然地继续聊天。",
		Goal:            "建立一段尊重边界、愿意认真投入的关系",
		TimeWindow:      "时间灵活",
		Location:        "清华校内公共场所",
		Visibility:      "candidates",
		Constraints:     []string{"尊重节奏", "有好奇心", "认真关系"},
		DealBreakers:    []string{"催促回复", "未经同意交换隐私"},
		DesiredNextStep: "双方愿意后先在校内散步半小时",
		PublishedAgo:    3 * time.Hour,
	},
}

func Ensure(ctx context.Context, dataStore *store.Store, cfg Config) (Result, error) {
	if len(cfg.Password) < 8 || len(cfg.Password) > 72 {
		return Result{}, errors.New("demo password must contain 8 to 72 bytes")
	}
	if dataStore == nil {
		return Result{}, errors.New("demo seed store is nil")
	}
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(cfg.Password),
		cfg.BCryptCost,
	)
	if err != nil {
		return Result{}, fmt.Errorf("hash demo password: %w", err)
	}

	now := time.Now().UTC()
	result := Result{}
	users := make(map[string]store.User, len(demoUsers))
	for _, seed := range demoUsers {
		user, created, err := ensureUser(ctx, dataStore, seed, string(passwordHash), now)
		if err != nil {
			return Result{}, err
		}
		if created {
			result.UsersCreated++
		}
		users[seed.Email] = user

		if err := ensureMatchingProfile(ctx, dataStore, user.ID, seed, now); err != nil {
			return Result{}, err
		}
	}

	intents := make(map[string]int64, len(demoIntents))
	for _, seed := range demoIntents {
		author, ok := users[seed.AuthorEmail]
		if !ok {
			return Result{}, fmt.Errorf("demo intent %q has no author", seed.Key)
		}
		intentID, created, err := ensureIntent(ctx, dataStore, author.ID, seed, now)
		if err != nil {
			return Result{}, err
		}
		if created {
			result.IntentsCreated++
		}
		intents[seed.Key] = intentID
	}

	primary := users[DemoEmail]
	projectCandidate := users["zhouyan@mails.tsinghua.edu.cn"]
	if err := ensurePendingInterest(
		ctx,
		dataStore,
		intents["primary-project"],
		projectCandidate.ID,
		"我做过几轮用户访谈和产品验证，也在找工程互补的长期伙伴。可以先看项目 memo，再坦诚聊投入和分工。",
		now,
	); err != nil {
		return Result{}, err
	}

	mealOwner := users["chenyu@mails.tsinghua.edu.cn"]
	roomCreated, err := ensureDemoRoom(
		ctx,
		dataStore,
		intents["meal"],
		mealOwner.ID,
		primary.ID,
		now,
	)
	if err != nil {
		return Result{}, err
	}
	if roomCreated {
		result.RoomsCreated++
	}

	return result, nil
}

func ensureUser(
	ctx context.Context,
	dataStore *store.Store,
	seed userSeed,
	passwordHash string,
	now time.Time,
) (store.User, bool, error) {
	user, err := dataStore.GetUserByEmail(ctx, nil, seed.Email)
	if errors.Is(err, store.ErrNotFound) {
		user, err = dataStore.CreateUser(ctx, nil, store.CreateUserParams{
			Email:        seed.Email,
			SchoolKey:    "thu",
			PasswordHash: passwordHash,
			Nickname:     seed.Nickname,
			Department:   seed.Department,
			Grade:        seed.Grade,
			Now:          now,
		})
		if err != nil {
			return store.User{}, false, fmt.Errorf("create demo user %s: %w", seed.Email, err)
		}
		user, err = dataStore.UpdateUserProfile(ctx, nil, store.UpdateProfileParams{
			UserID:       user.ID,
			Nickname:     user.Nickname,
			Department:   user.Department,
			Grade:        user.Grade,
			ContactType:  "wechat",
			ContactValue: seed.ContactValue,
			Now:          now,
		})
		if err != nil {
			return store.User{}, false, fmt.Errorf("set demo contact %s: %w", seed.Email, err)
		}
		return user, true, nil
	}
	if err != nil {
		return store.User{}, false, fmt.Errorf("find demo user %s: %w", seed.Email, err)
	}
	if err := dataStore.UpdateUserPassword(ctx, nil, user.ID, passwordHash, now); err != nil {
		return store.User{}, false, fmt.Errorf("refresh demo password %s: %w", seed.Email, err)
	}
	if user.ContactValue == "" {
		user, err = dataStore.UpdateUserProfile(ctx, nil, store.UpdateProfileParams{
			UserID:       user.ID,
			Nickname:     user.Nickname,
			Department:   user.Department,
			Grade:        user.Grade,
			ContactType:  "wechat",
			ContactValue: seed.ContactValue,
			Now:          now,
		})
		if err != nil {
			return store.User{}, false, fmt.Errorf("set existing demo contact %s: %w", seed.Email, err)
		}
	}
	return user, false, nil
}

func ensureMatchingProfile(
	ctx context.Context,
	dataStore *store.Store,
	userID int64,
	seed userSeed,
	now time.Time,
) error {
	if _, err := dataStore.GetMatchingProfile(ctx, nil, userID); err == nil {
		return nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("find demo matching profile: %w", err)
	}

	_, err := dataStore.UpsertMatchingProfile(ctx, nil, store.UpsertMatchingProfileParams{
		UserID:             userID,
		Bio:                seed.Bio,
		Skills:             seed.Skills,
		Interests:          seed.Interests,
		Availability:       seed.Availability,
		PreferredScenes:    seed.PreferredScenes,
		CollaborationStyle: seed.CollaborationStyle,
		Boundaries:         seed.Boundaries,
		AIPortrait:         seed.AIPortrait,
		AIPortraitApproved: true,
		RomanticOptIn:      seed.RomanticOptIn,
		MatchingPaused:     false,
		Now:                now,
	})
	if err != nil {
		return fmt.Errorf("create demo matching profile: %w", err)
	}
	return nil
}

func ensureIntent(
	ctx context.Context,
	dataStore *store.Store,
	authorID int64,
	seed intentSeed,
	now time.Time,
) (int64, bool, error) {
	intentID, err := dataStore.FindConnectionIntentIDByAuthorAndTitle(
		ctx,
		nil,
		authorID,
		seed.Title,
	)
	if err == nil {
		if seed.ProjectMemo.Building != "" {
			existing, getErr := dataStore.GetConnectionIntent(
				ctx,
				nil,
				authorID,
				intentID,
			)
			if getErr != nil {
				return 0, false, fmt.Errorf("read demo intent %q: %w", seed.Key, getErr)
			}
			if existing.ProjectMemo.Building == "" {
				existing.ProjectMemo = seed.ProjectMemo
				if updateErr := dataStore.UpdateConnectionIntent(
					ctx,
					nil,
					store.UpdateConnectionIntentParams{
						IntentID:        existing.ID,
						AuthorID:        existing.AuthorID,
						Category:        existing.Category,
						Title:           existing.Title,
						Description:     existing.Description,
						Goal:            existing.Goal,
						TimeWindow:      existing.TimeWindow,
						Location:        existing.Location,
						Visibility:      existing.Visibility,
						Status:          existing.Status,
						Constraints:     existing.Constraints,
						DealBreakers:    existing.DealBreakers,
						DesiredNextStep: existing.DesiredNextStep,
						ProjectMemo:     existing.ProjectMemo,
						ExpiresAt:       existing.ExpiresAt,
						PublishedAt:     existing.PublishedAt,
						Now:             now,
					},
				); updateErr != nil {
					return 0, false, fmt.Errorf(
						"backfill demo project memo %q: %w",
						seed.Key,
						updateErr,
					)
				}
			}
		}
		return intentID, false, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return 0, false, fmt.Errorf("find demo intent %q: %w", seed.Key, err)
	}

	publishedAt := now.Add(-seed.PublishedAgo)
	intentID, err = dataStore.CreateConnectionIntent(ctx, nil, store.CreateConnectionIntentParams{
		AuthorID:        authorID,
		Category:        seed.Category,
		Title:           seed.Title,
		Description:     seed.Description,
		Goal:            seed.Goal,
		TimeWindow:      seed.TimeWindow,
		Location:        seed.Location,
		Visibility:      seed.Visibility,
		Status:          "active",
		Constraints:     seed.Constraints,
		DealBreakers:    seed.DealBreakers,
		DesiredNextStep: seed.DesiredNextStep,
		ProjectMemo:     seed.ProjectMemo,
		PublishedAt:     sql.NullTime{Time: publishedAt, Valid: true},
		Now:             publishedAt,
	})
	if err != nil {
		return 0, false, fmt.Errorf("create demo intent %q: %w", seed.Key, err)
	}
	return intentID, true, nil
}

func ensurePendingInterest(
	ctx context.Context,
	dataStore *store.Store,
	intentID int64,
	candidateID int64,
	message string,
	now time.Time,
) error {
	if _, err := dataStore.GetIntentInterestStatus(ctx, nil, intentID, candidateID); err == nil {
		return nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("find demo interest: %w", err)
	}
	if err := dataStore.SetIntentInterest(
		ctx,
		nil,
		intentID,
		candidateID,
		"interested",
		message,
		now,
	); err != nil {
		return fmt.Errorf("create demo interest: %w", err)
	}
	return nil
}

func ensureDemoRoom(
	ctx context.Context,
	dataStore *store.Store,
	intentID int64,
	ownerID int64,
	candidateID int64,
	now time.Time,
) (bool, error) {
	if _, err := dataStore.FindMatchRoomID(ctx, nil, intentID, candidateID); err == nil {
		return false, nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return false, fmt.Errorf("find demo match room: %w", err)
	}

	if err := dataStore.SetIntentInterest(
		ctx,
		nil,
		intentID,
		candidateID,
		"accepted",
		"今晚有空，也想聊聊 AI 社交产品的实际难点。",
		now.Add(-25*time.Minute),
	); err != nil {
		return false, fmt.Errorf("create accepted demo interest: %w", err)
	}

	roomID, err := dataStore.CreateMatchRoom(ctx, nil, store.CreateMatchRoomParams{
		IntentID:      intentID,
		OwnerID:       ownerID,
		CandidateID:   candidateID,
		FitReason:     "你们都在做 AI 相关产品，也都偏好从具体问题开始聊天。双方已经明确愿意在今晚见面，这可能值得一次低压力的交流。",
		Uncertainty:   "你们对产品方向的判断未必一致；这正适合在第一次见面时用真实案例确认。",
		SharedContext: []string{"AI 产品", "今晚有空", "紫荆园"},
		ConversationStarters: []string{
			"你最近遇到的一个真正困难的 AI 产品问题是什么？",
			"什么样的社交产品指标最容易把团队带偏？",
			"如果只有一周，你会先验证哪个假设？",
		},
		SuggestedPlaces: []string{"紫荆园一层", "桃李园外侧座位", "近春园长椅"},
		Now:             now.Add(-20 * time.Minute),
	})
	if err != nil {
		return false, fmt.Errorf("create demo match room: %w", err)
	}

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("CST", 8*60*60)
	}
	localNow := now.In(location)
	scheduledAt := time.Date(
		localNow.Year(),
		localNow.Month(),
		localNow.Day()+1,
		19,
		0,
		0,
		0,
		location,
	).UTC()
	if err := dataStore.UpdateMatchRoomPlan(
		ctx,
		nil,
		ownerID,
		roomID,
		sql.NullTime{Time: scheduledAt, Valid: true},
		"紫荆园一层",
		now.Add(-15*time.Minute),
	); err != nil {
		return false, fmt.Errorf("plan demo match room: %w", err)
	}
	if err := dataStore.SetMatchRoomContactShared(
		ctx,
		nil,
		ownerID,
		roomID,
		true,
		now.Add(-14*time.Minute),
	); err != nil {
		return false, fmt.Errorf("share demo owner contact: %w", err)
	}
	if err := dataStore.SetMatchRoomContactShared(
		ctx,
		nil,
		candidateID,
		roomID,
		true,
		now.Add(-13*time.Minute),
	); err != nil {
		return false, fmt.Errorf("share demo candidate contact: %w", err)
	}
	if _, err := dataStore.CreateMatchRoomMessage(
		ctx,
		nil,
		roomID,
		ownerID,
		"你好，我把时间先放在明晚七点，紫荆园一层可以吗？",
		now.Add(-10*time.Minute),
	); err != nil {
		return false, fmt.Errorf("create first demo message: %w", err)
	}
	if _, err := dataStore.CreateMatchRoomMessage(
		ctx,
		nil,
		roomID,
		candidateID,
		"可以，我会带上目前的产品原型，我们从各自遇到的真实问题聊起。",
		now.Add(-8*time.Minute),
	); err != nil {
		return false, fmt.Errorf("create second demo message: %w", err)
	}
	return true, nil
}

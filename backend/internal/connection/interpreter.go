package connection

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ClarifyingQuestion struct {
	ID      string   `json:"id"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"`
}

type SafetyAssessment struct {
	Level  string `json:"level"`
	Notice string `json:"notice"`
}

type IntentInterpretation struct {
	Category        string               `json:"category"`
	Title           string               `json:"title"`
	Description     string               `json:"description"`
	Goal            string               `json:"goal"`
	TimeWindow      string               `json:"timeWindow"`
	Location        string               `json:"location"`
	Visibility      string               `json:"visibility"`
	Constraints     []string             `json:"constraints"`
	DesiredNextStep string               `json:"desiredNextStep"`
	Confidence      float64              `json:"confidence"`
	Questions       []ClarifyingQuestion `json:"questions"`
	Inferences      []string             `json:"inferences"`
	Safety          SafetyAssessment     `json:"safety"`
	Engine          string               `json:"engine"`
}

type Interpreter struct{}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (i *Interpreter) Interpret(raw string) IntentInterpretation {
	description := strings.TrimSpace(raw)
	category, categoryConfidence := inferCategory(description)
	location := inferLocation(description)
	timeWindow := inferTimeWindow(description)
	constraints := inferConstraints(description)
	visibility := "feed"
	if category == CategoryRomantic {
		visibility = "candidates"
	}

	questions := make([]ClarifyingQuestion, 0, 3)
	if category == CategoryOther {
		questions = append(questions, ClarifyingQuestion{
			ID:      "category",
			Prompt:  "这次见面最想解决什么事情？",
			Options: []string{"一起吃饭或活动", "学习或项目协作", "认真认识一个人"},
		})
	}
	if timeWindow == "" {
		questions = append(questions, ClarifyingQuestion{
			ID:      "timeWindow",
			Prompt:  "你希望大约什么时候见面或开始聊？",
			Options: []string{"今天", "本周", "时间都可以"},
		})
	}
	if category == CategoryProject || category == CategoryCofounder {
		if !containsAny(description, []string{"前端", "后端", "算法", "产品经理", "设计", "运营", "全栈", "infra", "工程"}) {
			questions = append(questions, ClarifyingQuestion{
				ID:      "skills",
				Prompt:  "对方最需要具备哪类能力？",
				Options: []string{"技术研发", "产品与设计", "增长与运营"},
			})
		}
	} else if location == "" && len(questions) < 3 {
		questions = append(questions, ClarifyingQuestion{
			ID:      "location",
			Prompt:  "更方便在哪个校内区域碰面？",
			Options: []string{"紫荆附近", "图书馆附近", "校内都可以"},
		})
	}
	if len(questions) > 3 {
		questions = questions[:3]
	}

	inferences := []string{"意图类别由描述中的场景词推断"}
	if timeWindow != "" {
		inferences = append(inferences, "时间窗口由描述中的时间表达推断")
	}
	if location != "" {
		inferences = append(inferences, "地点偏好由校内地点名称推断")
	}
	if category == CategoryRomantic {
		inferences = append(inferences, "恋爱意图默认仅向推荐候选人可见")
	}

	confidence := categoryConfidence
	if timeWindow != "" {
		confidence += 0.06
	}
	if location != "" {
		confidence += 0.05
	}
	if confidence > 0.96 {
		confidence = 0.96
	}

	return IntentInterpretation{
		Category:        category,
		Title:           inferredTitle(category, location),
		Description:     description,
		Goal:            inferredGoal(category),
		TimeWindow:      timeWindow,
		Location:        location,
		Visibility:      visibility,
		Constraints:     constraints,
		DesiredNextStep: inferredNextStep(category),
		Confidence:      confidence,
		Questions:       questions,
		Inferences:      inferences,
		Safety:          assessSafety(description),
		Engine:          "soulink-intent-engine-v1",
	}
}

func inferCategory(text string) (string, float64) {
	switch {
	case containsAny(text, []string{"cto", "联合创始", "合伙人", "cofounder", "创业搭档"}):
		return CategoryCofounder, 0.9
	case containsAny(text, []string{"对象", "恋爱", "脱单", "约会", "男朋友", "女朋友", "伴侣", "认真认识"}):
		return CategoryRomantic, 0.9
	case containsAny(text, []string{"项目", "组队", "队友", "前端", "后端", "算法", "产品经理", "设计师", "黑客松", "比赛"}):
		return CategoryProject, 0.86
	case containsAny(text, []string{"吃饭", "饭搭子", "晚饭", "午饭", "早餐", "食堂", "夜宵"}):
		return CategoryMeal, 0.92
	case containsAny(text, []string{"自习", "学习", "刷题", "备考", "图书馆", "作业"}):
		return CategoryStudy, 0.88
	case containsAny(text, []string{"跑步", "羽毛球", "网球", "游泳", "健身", "运动", "骑行", "飞盘", "爬山"}):
		return CategorySport, 0.88
	case containsAny(text, []string{"讲座", "演出", "展览", "活动", "音乐会", "电影", "演讲"}):
		return CategoryEvent, 0.82
	case containsAny(text, []string{"聊聊", "聊天", "倾诉", "散步", "认识新朋友", "交流"}):
		return CategoryConversation, 0.78
	default:
		return CategoryOther, 0.48
	}
}

func inferLocation(text string) string {
	locations := []string{
		"紫荆园", "紫荆", "桃李园", "桃李", "听涛园", "听涛", "清芬园", "清芬",
		"观畴园", "观畴", "图书馆北馆", "图书馆", "六教", "三教", "一教",
		"东操", "西操", "综体", "情人坡", "近春园", "清青", "FIT楼",
	}
	for _, location := range locations {
		if strings.Contains(strings.ToLower(text), strings.ToLower(location)) {
			return location
		}
	}
	return ""
}

func inferTimeWindow(text string) string {
	switch {
	case containsAny(text, []string{"今晚", "今天晚上", "今夜"}):
		return "今晚"
	case containsAny(text, []string{"今天中午", "中午"}):
		return "今天中午"
	case strings.Contains(text, "今天"):
		return "今天"
	case containsAny(text, []string{"明晚", "明天晚上"}):
		return "明晚"
	case strings.Contains(text, "明天"):
		return "明天"
	case containsAny(text, []string{"这周末", "本周末", "周末"}):
		return "本周末"
	case containsAny(text, []string{"这周", "本周"}):
		return "本周"
	case containsAny(text, []string{"长期", "每周", "固定"}):
		return "长期固定"
	default:
		return ""
	}
}

func inferConstraints(text string) []string {
	candidates := []string{
		"安静", "认真", "长期", "高效", "低压力", "守时", "有经验", "技术", "线下",
		"每周", "晚上", "周末", "女生", "男生",
	}
	result := make([]string, 0, 4)
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(text), strings.ToLower(candidate)) {
			result = append(result, candidate)
		}
		if len(result) == 4 {
			break
		}
	}
	return result
}

func inferredTitle(category, location string) string {
	titles := map[string]string{
		CategoryMeal:         "找一位饭搭子",
		CategoryStudy:        "找一位学习搭子",
		CategorySport:        "找一位运动搭子",
		CategoryEvent:        "一起参加校内活动",
		CategoryConversation: "找一个可以认真聊聊的人",
		CategoryRomantic:     "认真认识一个人",
		CategoryProject:      "寻找项目伙伴",
		CategoryCofounder:    "寻找 CTO 或联合创始人",
		CategoryOther:        "认识一个值得见面的人",
	}
	title := titles[category]
	if location != "" && utf8.RuneCountInString(title+" · "+location) <= 36 {
		title += " · " + location
	}
	return title
}

func inferredGoal(category string) string {
	goals := map[string]string{
		CategoryMeal:         "轻松吃顿饭，认识一个聊得来的人",
		CategoryStudy:        "建立一次或持续的共同学习节奏",
		CategorySport:        "找到时间和强度合适的运动伙伴",
		CategoryEvent:        "一起参加活动，减少一个人出发的门槛",
		CategoryConversation: "进行一次真诚、低压力的交流",
		CategoryRomantic:     "在彼此愿意的前提下认真了解对方",
		CategoryProject:      "找到技能和投入程度合适的协作者",
		CategoryCofounder:    "找到愿景、能力和承诺匹配的长期伙伴",
		CategoryOther:        "把一个模糊的找人需求变成一次真实见面",
	}
	return goals[category]
}

func inferredNextStep(category string) string {
	switch category {
	case CategoryProject, CategoryCofounder:
		return "先进行 30 分钟项目交流"
	case CategoryRomantic:
		return "双方同意后先轻松聊聊"
	default:
		return "确认时间地点后线下见面"
	}
}

func assessSafety(text string) SafetyAssessment {
	compact := strings.ReplaceAll(strings.TrimSpace(text), " ", "")
	if compact == "" {
		return SafetyAssessment{Level: "review", Notice: "请先描述你想认识什么样的人"}
	}

	phonePattern := regexp.MustCompile(`1[3-9][0-9]{9}`)
	if phonePattern.MatchString(compact) || strings.Contains(strings.ToLower(compact), "微信号") {
		return SafetyAssessment{
			Level:  "review",
			Notice: "描述里可能包含联系方式；发布前建议删除，双方同意后再在匹配房间中揭示",
		}
	}
	if address, err := mail.ParseAddress(strings.TrimSpace(text)); err == nil && strings.Contains(address.Address, "@") {
		return SafetyAssessment{
			Level:  "review",
			Notice: "描述里可能包含邮箱；发布前建议删除，双方同意后再交换联系方式",
		}
	}
	return SafetyAssessment{Level: "clear", Notice: "未发现需要额外确认的公开信息"}
}

func containsAny(text string, values []string) bool {
	normalized := strings.ToLower(text)
	for _, value := range values {
		if strings.Contains(normalized, strings.ToLower(value)) {
			return true
		}
	}
	return false
}

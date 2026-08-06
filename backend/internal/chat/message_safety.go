package chat

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"soulink-backend/internal/ai"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

var (
	messageURLPattern        = regexp.MustCompile(`(?i)https?://|www\.|[a-z0-9-]+\.(?:com|cn|net|top)(?:/|\b)`)
	messageSpamPattern       = regexp.MustCompile(`(?i)(?:兼职|返利|推广|代写|出售|扫码|赚钱|拉群|进群|加群|内部渠道|稳赚)`)
	messageHarassmentPattern = regexp.MustCompile(`(?:傻逼|废物|垃圾人|滚开|贱人|恶心死了|操你)`)
	messageThreatPattern     = regexp.MustCompile(`(?:杀了你|弄死你|堵你|跟踪你|收拾你|让你付出代价|你给我等着)`)
	messagePrivacyPattern    = regexp.MustCompile(`(?:开盒|曝光你|公布你的|发你的照片|人肉你|查你住哪)`)
	messageCoercionPattern   = regexp.MustCompile(`(?:不来就|必须单独|不许告诉|敢拒绝|必须见我|逼你)`)
)

var messageRiskCategories = map[string]struct{}{
	"spam": {}, "contact_inducement": {}, "harassment": {},
	"threat": {}, "privacy": {}, "coercion": {},
}

type MessageSafetyView struct {
	Level                string   `json:"level"`
	Categories           []string `json:"categories"`
	Reasons              []string `json:"reasons"`
	RequiresConfirmation bool     `json:"requiresConfirmation"`
	InferenceSource      string   `json:"inferenceSource"`
	InferenceModel       string   `json:"inferenceModel,omitempty"`
	Confidence           float64  `json:"confidence"`
	PromptVersion        string   `json:"promptVersion"`
}

type messageSafetyModelOutput struct {
	Level      string   `json:"level"`
	Categories []string `json:"categories"`
	Reasons    []string `json:"reasons"`
	Confidence float64  `json:"confidence"`
}

func (s *Service) CheckDirectMessageSafety(
	ctx context.Context,
	userID, conversationID int64,
	content string,
) (MessageSafetyView, error) {
	conversation, err := s.store.GetConversation(ctx, nil, conversationID)
	if err != nil {
		return MessageSafetyView{}, mapMessageSafetyContextError(err, "会话不存在")
	}
	if conversation.User1ID != userID && conversation.User2ID != userID {
		return MessageSafetyView{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权访问该会话")
	}
	return s.checkMessageSafety(ctx, content)
}

func (s *Service) CheckGroupMessageSafety(
	ctx context.Context,
	userID, groupID int64,
	content string,
) (MessageSafetyView, error) {
	if _, err := s.requireGroupMember(ctx, groupID, userID); err != nil {
		return MessageSafetyView{}, err
	}
	return s.checkMessageSafety(ctx, content)
}

func (s *Service) checkMessageSafety(ctx context.Context, content string) (MessageSafetyView, error) {
	content = strings.TrimSpace(content)
	if content == "" || utf8.RuneCountInString(content) > maxChatMessageLength {
		return MessageSafetyView{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "消息内容不符合要求")
	}
	prompt, _ := ai.PromptFor(ai.TaskMessageSafety)
	ruleResult := deterministicMessageSafety(content)
	ruleResult.PromptVersion = prompt.Version
	if s.generator == nil {
		return ruleResult, nil
	}
	modelInput := map[string]any{
		"redactedText": redactAssistantContacts(content),
		"ruleLevel":    ruleResult.Level,
		"ruleSignals":  ruleResult.Categories,
	}
	var output messageSafetyModelOutput
	metadata, err := s.generator.Generate(ctx, prompt, modelInput, &output)
	if err != nil || !validMessageSafetyModelOutput(output) {
		return ruleResult, nil
	}
	merged := mergeMessageSafety(ruleResult, output)
	merged.InferenceSource = metadata.Source
	merged.InferenceModel = metadata.Model
	merged.PromptVersion = metadata.PromptVersion
	return merged, nil
}

func deterministicMessageSafety(content string) MessageSafetyView {
	categories := make([]string, 0, 4)
	reasons := make([]string, 0, 4)
	level := "clear"
	add := func(category, reason, nextLevel string) {
		if !containsMessageSafetyValue(categories, category) {
			categories = append(categories, category)
			reasons = append(reasons, reason)
		}
		if messageSafetySeverity(nextLevel) > messageSafetySeverity(level) {
			level = nextLevel
		}
	}
	if assistantEmailPattern.MatchString(content) || assistantPhonePattern.MatchString(content) || assistantContactPattern.MatchString(content) {
		add("contact_inducement", "消息包含站外联系方式，请确认是否已经建立足够信任。", "caution")
	}
	if messageSpamPattern.MatchString(content) || messageURLPattern.MatchString(content) || hasRepeatedMessageRun(content, 8) {
		add("spam", "消息呈现推广、外链或重复发送特征。", "caution")
	}
	if messageHarassmentPattern.MatchString(content) {
		add("harassment", "消息包含可能伤害对方的侮辱或骚扰表达。", "high")
	}
	if messageThreatPattern.MatchString(content) {
		add("threat", "消息包含威胁或可能升级冲突的表达。", "high")
	}
	if messagePrivacyPattern.MatchString(content) {
		add("privacy", "消息涉及曝光、追查或传播他人隐私。", "high")
	}
	if messageCoercionPattern.MatchString(content) {
		add("coercion", "消息包含强迫见面或限制求助的表达。", "high")
	}
	confidence := 0.72
	if level == "clear" {
		confidence = 0.66
	} else if level == "high" {
		confidence = 0.9
	}
	return MessageSafetyView{
		Level: level, Categories: categories, Reasons: reasons,
		RequiresConfirmation: level == "high", InferenceSource: "deterministic",
		Confidence: confidence,
	}
}

func hasRepeatedMessageRun(value string, limit int) bool {
	var previous rune
	count := 0
	for _, current := range value {
		if current == previous {
			count++
		} else {
			previous = current
			count = 1
		}
		if count >= limit {
			return true
		}
	}
	return false
}

func mergeMessageSafety(rule MessageSafetyView, model messageSafetyModelOutput) MessageSafetyView {
	result := rule
	if messageSafetySeverity(model.Level) > messageSafetySeverity(result.Level) {
		result.Level = model.Level
	}
	for index, category := range model.Categories {
		if containsMessageSafetyValue(result.Categories, category) {
			continue
		}
		result.Categories = append(result.Categories, category)
		if index < len(model.Reasons) && strings.TrimSpace(model.Reasons[index]) != "" {
			result.Reasons = append(result.Reasons, assistantTruncate(model.Reasons[index], 160))
		}
	}
	if len(result.Reasons) > 6 {
		result.Reasons = result.Reasons[:6]
	}
	if model.Confidence > result.Confidence {
		result.Confidence = model.Confidence
	}
	result.RequiresConfirmation = result.Level == "high"
	return result
}

func validMessageSafetyModelOutput(output messageSafetyModelOutput) bool {
	if messageSafetySeverity(output.Level) < 0 || output.Confidence < 0 || output.Confidence > 1 ||
		len(output.Categories) > 6 || len(output.Reasons) > 6 {
		return false
	}
	for _, category := range output.Categories {
		if _, ok := messageRiskCategories[category]; !ok {
			return false
		}
	}
	return true
}

func messageSafetySeverity(level string) int {
	switch level {
	case "clear":
		return 0
	case "caution":
		return 1
	case "high":
		return 2
	default:
		return -1
	}
}

func containsMessageSafetyValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func mapMessageSafetyContextError(err error, message string) error {
	if err == store.ErrNotFound {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, message)
	}
	return httpapi.InternalError(err)
}

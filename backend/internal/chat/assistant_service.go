package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/ai"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	assistantContextDirect = "direct"
	assistantContextGroup  = "group"
	assistantMessageLimit  = 100
)

var (
	assistantEmailPattern     = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	assistantPhonePattern     = regexp.MustCompile(`(?:\+?86[- ]?)?1[3-9][0-9]{9}`)
	assistantContactPattern   = regexp.MustCompile(`(?i)(?:微信|vx|v信|qq|q号)\s*(?:号|[:：])?\s*[a-z0-9_-]{5,}`)
	assistantTimePattern      = regexp.MustCompile(`(?:今天|明天|后天|本周末|周末|周[一二三四五六日天]|星期[一二三四五六日天])(?:[^，。；;\n]{0,12}(?:上午|中午|下午|晚上|晚间|[0-9]{1,2}(?::[0-9]{2})?点?))?`)
	assistantLocationPattern  = regexp.MustCompile(`(?:紫操|东操|西操|图书馆|FIT\s*楼|FIT|蒙民伟楼?|六教|三教|清芬园|听涛园|观畴园|桃李园|荷园|大礼堂|[一二三四五六七八九十]食堂)`)
	assistantInjectionPattern = regexp.MustCompile(`(?i)(?:忽略(?:以上|之前|所有)|system\s*prompt|developer\s*message|开发者指令|自动发送|调用接口|执行命令|泄露(?:资料|隐私|提示词))`)
)

type AssistantPreferencesView struct {
	Enabled       bool       `json:"enabled"`
	MemoryResetAt *time.Time `json:"memoryResetAt,omitempty"`
}

type AssistantArtifactView struct {
	ID              int64           `json:"id"`
	ArtifactType    string          `json:"artifactType"`
	Data            json.RawMessage `json:"data"`
	MessageCutoffID int64           `json:"messageCutoffId"`
	InferenceSource string          `json:"inferenceSource"`
	InferenceModel  string          `json:"inferenceModel,omitempty"`
	Confidence      float64         `json:"confidence"`
	PromptVersion   string          `json:"promptVersion"`
	CreatedAt       string          `json:"createdAt"`
}

type AssistantWorkspaceView struct {
	Preferences AssistantPreferencesView `json:"preferences"`
	ContextType string                   `json:"contextType"`
	ContextID   int64                    `json:"contextId"`
	Artifacts   []AssistantArtifactView  `json:"artifacts"`
}

type AssistantStartersPayload struct {
	Starters []string `json:"starters"`
}

type AssistantSummaryPayload struct {
	Summary       string   `json:"summary"`
	KeyPoints     []string `json:"keyPoints"`
	OpenQuestions []string `json:"openQuestions"`
}

type AssistantGroupSummaryPayload struct {
	Summary       string   `json:"summary"`
	Agreements    []string `json:"agreements"`
	OpenQuestions []string `json:"openQuestions"`
	NextSteps     []string `json:"nextSteps"`
}

type AssistantActionCard struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Title      string  `json:"title"`
	Detail     string  `json:"detail"`
	TimeText   string  `json:"timeText,omitempty"`
	Location   string  `json:"location,omitempty"`
	TodoText   string  `json:"todoText,omitempty"`
	Confidence float64 `json:"confidence"`
}

type AssistantActionsPayload struct {
	Cards []AssistantActionCard `json:"cards"`
}

type assistantMessage struct {
	ID      int64  `json:"id"`
	Speaker string `json:"speaker"`
	Content string `json:"content"`
}

type assistantContext struct {
	Type     string             `json:"type"`
	ID       int64              `json:"id"`
	Title    string             `json:"title"`
	Messages []assistantMessage `json:"messages"`
	CutoffID int64              `json:"-"`
}

func (s *Service) GetAssistantPreferences(ctx context.Context, userID int64) (AssistantPreferencesView, error) {
	preferences, err := s.store.GetAIUserPreferences(ctx, nil, userID)
	if errors.Is(err, store.ErrNotFound) {
		return AssistantPreferencesView{Enabled: true}, nil
	}
	if err != nil {
		return AssistantPreferencesView{}, httpapi.InternalError(err)
	}
	return assistantPreferencesView(preferences), nil
}

func (s *Service) SetAssistantEnabled(ctx context.Context, userID int64, enabled bool) (AssistantPreferencesView, error) {
	preferences, err := s.store.SetAssistantEnabled(ctx, nil, userID, enabled, s.now())
	if err != nil {
		return AssistantPreferencesView{}, httpapi.InternalError(err)
	}
	return assistantPreferencesView(preferences), nil
}

func (s *Service) ResetAssistantMemory(ctx context.Context, userID int64) (AssistantPreferencesView, error) {
	now := s.now()
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.store.ResetAssistantMemory(ctx, tx, userID, now); err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	}); err != nil {
		return AssistantPreferencesView{}, err
	}
	return s.GetAssistantPreferences(ctx, userID)
}

func (s *Service) GetAssistantWorkspace(
	ctx context.Context,
	userID int64,
	contextType string,
	contextID int64,
) (AssistantWorkspaceView, error) {
	if _, err := s.loadAssistantContext(ctx, userID, contextType, contextID); err != nil {
		return AssistantWorkspaceView{}, err
	}
	preferences, err := s.GetAssistantPreferences(ctx, userID)
	if err != nil {
		return AssistantWorkspaceView{}, err
	}
	artifacts, err := s.store.ListLatestChatAssistantArtifacts(ctx, nil, userID, contextType, contextID)
	if err != nil {
		return AssistantWorkspaceView{}, httpapi.InternalError(err)
	}
	views := make([]AssistantArtifactView, 0, len(artifacts))
	for _, artifact := range artifacts {
		views = append(views, assistantArtifactView(artifact))
	}
	return AssistantWorkspaceView{
		Preferences: preferences, ContextType: contextType,
		ContextID: contextID, Artifacts: views,
	}, nil
}

func (s *Service) GenerateAssistantArtifact(
	ctx context.Context,
	userID int64,
	contextType string,
	contextID int64,
	requestedType string,
) (AssistantArtifactView, error) {
	preferences, err := s.GetAssistantPreferences(ctx, userID)
	if err != nil {
		return AssistantArtifactView{}, err
	}
	if !preferences.Enabled {
		return AssistantArtifactView{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "聊天助手已关闭")
	}
	conversationContext, err := s.loadAssistantContext(ctx, userID, contextType, contextID)
	if err != nil {
		return AssistantArtifactView{}, err
	}

	artifactType := strings.TrimSpace(requestedType)
	var payload any
	var metadata ai.Metadata
	var confidence float64
	switch artifactType {
	case "starters":
		if contextType != assistantContextDirect {
			return AssistantArtifactView{}, invalidAssistantRequest("开场建议仅用于一对一会话")
		}
		payload, metadata, confidence = s.generateAssistantStarters(ctx, conversationContext)
	case "summary":
		if contextType == assistantContextGroup {
			artifactType = "group_summary"
			payload, metadata, confidence = s.generateAssistantGroupSummary(ctx, conversationContext)
		} else {
			payload, metadata, confidence = s.generateAssistantSummary(ctx, conversationContext)
		}
	case "actions":
		payload, metadata, confidence = s.generateAssistantActions(ctx, conversationContext)
	default:
		return AssistantArtifactView{}, invalidAssistantRequest("不支持的助手操作")
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return AssistantArtifactView{}, httpapi.InternalError(err)
	}
	artifact, err := s.store.CreateChatAssistantArtifact(ctx, nil, store.CreateChatAssistantArtifactParams{
		UserID: userID, ContextType: contextType, ContextID: contextID,
		ArtifactType: artifactType, Payload: rawPayload,
		MessageCutoffID: conversationContext.CutoffID,
		InferenceSource: metadata.Source, InferenceModel: metadata.Model,
		Confidence: confidence, PromptVersion: metadata.PromptVersion, Now: s.now(),
	})
	if err != nil {
		return AssistantArtifactView{}, httpapi.InternalError(err)
	}
	return assistantArtifactView(artifact), nil
}

func (s *Service) DeleteAssistantArtifact(
	ctx context.Context,
	userID int64,
	contextType string,
	contextID, artifactID int64,
) error {
	if _, err := s.loadAssistantContext(ctx, userID, contextType, contextID); err != nil {
		return err
	}
	if err := s.store.SoftDeleteChatAssistantArtifact(
		ctx, nil, userID, contextType, contextID, artifactID, s.now(),
	); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "助手草稿不存在")
		}
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) loadAssistantContext(
	ctx context.Context,
	userID int64,
	contextType string,
	contextID int64,
) (assistantContext, error) {
	if contextID <= 0 || (contextType != assistantContextDirect && contextType != assistantContextGroup) {
		return assistantContext{}, invalidAssistantRequest("会话参数错误")
	}
	result := assistantContext{Type: contextType, ID: contextID, Messages: []assistantMessage{}}
	if contextType == assistantContextDirect {
		conversation, err := s.store.GetConversation(ctx, nil, contextID)
		if err != nil {
			return assistantContext{}, mapAssistantContextError(err, "会话不存在")
		}
		if conversation.User1ID != userID && conversation.User2ID != userID {
			return assistantContext{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权访问该会话")
		}
		peerID := conversation.User1ID
		if peerID == userID {
			peerID = conversation.User2ID
		}
		peer, err := s.store.GetUserByID(ctx, nil, peerID)
		if err != nil {
			return assistantContext{}, httpapi.InternalError(err)
		}
		result.Title = "和 " + peer.Nickname + " 的对话"
		messages, err := s.store.ListMessages(ctx, nil, contextID, 0, assistantMessageLimit)
		if err != nil {
			return assistantContext{}, httpapi.InternalError(err)
		}
		for index := len(messages) - 1; index >= 0; index-- {
			message := messages[index]
			speaker := "对方"
			if message.SenderID == userID {
				speaker = "我"
			}
			result.Messages = append(result.Messages, assistantMessage{
				ID: message.ID, Speaker: speaker,
				Content: sanitizeAssistantInput(message.Content),
			})
			if message.ID > result.CutoffID {
				result.CutoffID = message.ID
			}
		}
		return result, nil
	}

	group, err := s.requireGroupMember(ctx, contextID, userID)
	if err != nil {
		return assistantContext{}, err
	}
	result.Title = group.Name
	messages, err := s.store.ListChatGroupMessages(ctx, nil, contextID, 0, assistantMessageLimit)
	if err != nil {
		return assistantContext{}, httpapi.InternalError(err)
	}
	for index := len(messages) - 1; index >= 0; index-- {
		message := messages[index]
		speaker := message.SenderNickname
		if message.SenderID == userID {
			speaker = "我"
		}
		result.Messages = append(result.Messages, assistantMessage{
			ID: message.ID, Speaker: speaker,
			Content: sanitizeAssistantInput(message.Content),
		})
		if message.ID > result.CutoffID {
			result.CutoffID = message.ID
		}
	}
	return result, nil
}

func (s *Service) generateAssistantStarters(
	ctx context.Context,
	conversation assistantContext,
) (AssistantStartersPayload, ai.Metadata, float64) {
	prompt, _ := ai.PromptFor(ai.TaskConversationStarters)
	fallback := deterministicAssistantStarters(conversation)
	if s.generator == nil {
		return fallback, deterministicAssistantMetadata(prompt), 0.72
	}
	var output AssistantStartersPayload
	metadata, err := s.generator.Generate(ctx, prompt, conversation, &output)
	if err != nil || !validAssistantStarters(output) {
		return fallback, deterministicAssistantMetadata(prompt), 0.68
	}
	for index := range output.Starters {
		output.Starters[index] = sanitizeAssistantOutput(output.Starters[index])
	}
	return output, metadata, 0.84
}

func (s *Service) generateAssistantSummary(
	ctx context.Context,
	conversation assistantContext,
) (AssistantSummaryPayload, ai.Metadata, float64) {
	prompt, _ := ai.PromptFor(ai.TaskChatSummary)
	fallback := deterministicAssistantSummary(conversation)
	if s.generator == nil || len(conversation.Messages) == 0 {
		confidence := 0.68
		if len(conversation.Messages) == 0 {
			confidence = 0.32
		}
		return fallback, deterministicAssistantMetadata(prompt), confidence
	}
	var output AssistantSummaryPayload
	metadata, err := s.generator.Generate(ctx, prompt, conversation, &output)
	if err != nil || !validAssistantSummary(output) {
		return fallback, deterministicAssistantMetadata(prompt), 0.64
	}
	output = sanitizeAssistantSummary(output)
	return output, metadata, 0.82
}

func (s *Service) generateAssistantGroupSummary(
	ctx context.Context,
	conversation assistantContext,
) (AssistantGroupSummaryPayload, ai.Metadata, float64) {
	prompt, _ := ai.PromptFor(ai.TaskGroupSummary)
	fallback := deterministicAssistantGroupSummary(conversation)
	if s.generator == nil || len(conversation.Messages) == 0 {
		confidence := 0.7
		if len(conversation.Messages) == 0 {
			confidence = 0.32
		}
		return fallback, deterministicAssistantMetadata(prompt), confidence
	}
	var output AssistantGroupSummaryPayload
	metadata, err := s.generator.Generate(ctx, prompt, conversation, &output)
	if err != nil || !validAssistantGroupSummary(output) {
		return fallback, deterministicAssistantMetadata(prompt), 0.66
	}
	output = sanitizeAssistantGroupSummary(output)
	return output, metadata, 0.83
}

func (s *Service) generateAssistantActions(
	ctx context.Context,
	conversation assistantContext,
) (AssistantActionsPayload, ai.Metadata, float64) {
	prompt, _ := ai.PromptFor(ai.TaskChatActions)
	fallback := deterministicAssistantActions(conversation)
	if s.generator == nil || len(conversation.Messages) == 0 {
		return fallback, deterministicAssistantMetadata(prompt), assistantActionsConfidence(fallback)
	}
	var modelOutput struct {
		Cards []struct {
			Type       string  `json:"type"`
			Title      string  `json:"title"`
			Detail     string  `json:"detail"`
			TimeText   string  `json:"timeText"`
			Location   string  `json:"location"`
			TodoText   string  `json:"todoText"`
			Confidence float64 `json:"confidence"`
		} `json:"cards"`
	}
	metadata, err := s.generator.Generate(ctx, prompt, conversation, &modelOutput)
	if err != nil || len(modelOutput.Cards) > 8 {
		return fallback, deterministicAssistantMetadata(prompt), assistantActionsConfidence(fallback)
	}
	output := AssistantActionsPayload{Cards: make([]AssistantActionCard, 0, len(modelOutput.Cards))}
	for index, card := range modelOutput.Cards {
		if !validAssistantActionType(card.Type) || card.Confidence < 0 || card.Confidence > 1 {
			return fallback, deterministicAssistantMetadata(prompt), assistantActionsConfidence(fallback)
		}
		output.Cards = append(output.Cards, AssistantActionCard{
			ID: fmt.Sprintf("%s-%d", card.Type, index+1), Type: card.Type,
			Title: sanitizeAssistantOutput(card.Title), Detail: sanitizeAssistantOutput(card.Detail),
			TimeText: sanitizeAssistantOutput(card.TimeText), Location: sanitizeAssistantOutput(card.Location),
			TodoText: sanitizeAssistantOutput(card.TodoText), Confidence: card.Confidence,
		})
	}
	return output, metadata, assistantActionsConfidence(output)
}

func deterministicAssistantStarters(conversation assistantContext) AssistantStartersPayload {
	starters := []string{
		"你好，很高兴在 SouLink 认识你。最近最想先聊哪件事？",
		"我们可以先对齐一下彼此现在最舒服的交流节奏。",
		"如果你愿意，可以从最近正在投入的一件事聊起。",
	}
	if len(conversation.Messages) > 0 {
		last := assistantTruncate(conversation.Messages[len(conversation.Messages)-1].Content, 28)
		if last != "" && last != "[已忽略不安全指令]" {
			starters[0] = fmt.Sprintf("关于“%s”，我也想听听你的想法。", last)
		}
	}
	return AssistantStartersPayload{Starters: starters}
}

func deterministicAssistantSummary(conversation assistantContext) AssistantSummaryPayload {
	if len(conversation.Messages) == 0 {
		return AssistantSummaryPayload{
			Summary:   "当前还没有足够的对话内容可总结。",
			KeyPoints: []string{}, OpenQuestions: []string{},
		}
	}
	recent := assistantRecentMessages(conversation.Messages, 12)
	keyPoints := uniqueAssistantSnippets(recent, 3, func(message assistantMessage) bool {
		return message.Content != "[已忽略不安全指令]"
	})
	questions := uniqueAssistantSnippets(recent, 3, func(message assistantMessage) bool {
		return strings.ContainsAny(message.Content, "?？") || strings.Contains(message.Content, "吗")
	})
	last := assistantTruncate(recent[len(recent)-1].Content, 36)
	return AssistantSummaryPayload{
		Summary:   fmt.Sprintf("最近的对话围绕“%s”展开，仍以双方后续确认的信息为准。", last),
		KeyPoints: keyPoints, OpenQuestions: questions,
	}
}

func deterministicAssistantGroupSummary(conversation assistantContext) AssistantGroupSummaryPayload {
	if len(conversation.Messages) == 0 {
		return AssistantGroupSummaryPayload{
			Summary:    "群聊中还没有足够的消息可整理。",
			Agreements: []string{}, OpenQuestions: []string{}, NextSteps: []string{},
		}
	}
	recent := assistantRecentMessages(conversation.Messages, 20)
	agreements := uniqueAssistantSnippets(recent, 4, func(message assistantMessage) bool {
		return strings.ContainsAny(message.Content, "好可以行收到") || strings.Contains(message.Content, "确定")
	})
	questions := uniqueAssistantSnippets(recent, 4, func(message assistantMessage) bool {
		return strings.ContainsAny(message.Content, "?？") || strings.Contains(message.Content, "吗")
	})
	nextSteps := uniqueAssistantSnippets(recent, 4, func(message assistantMessage) bool {
		return strings.Contains(message.Content, "我来") || strings.Contains(message.Content, "你来") ||
			strings.Contains(message.Content, "需要") || strings.Contains(message.Content, "记得")
	})
	return AssistantGroupSummaryPayload{
		Summary:    fmt.Sprintf("%s 最近有 %d 条消息，以下仅整理明确表达的共识与待确认事项。", conversation.Title, len(recent)),
		Agreements: agreements, OpenQuestions: questions, NextSteps: nextSteps,
	}
}

func deterministicAssistantActions(conversation assistantContext) AssistantActionsPayload {
	cards := make([]AssistantActionCard, 0, 4)
	recent := assistantRecentMessages(conversation.Messages, 20)
	if value := latestAssistantPatternMatch(recent, assistantTimePattern); value != "" {
		cards = append(cards, AssistantActionCard{
			ID: "time-1", Type: "time", Title: "时间线索", Detail: value,
			TimeText: value, Confidence: 0.78,
		})
	}
	if value := latestAssistantPatternMatch(recent, assistantLocationPattern); value != "" {
		cards = append(cards, AssistantActionCard{
			ID: "location-1", Type: "location", Title: "地点线索", Detail: value,
			Location: value, Confidence: 0.8,
		})
	}
	if todo := latestAssistantSentenceContaining(recent, []string{"我来", "你来", "需要", "记得", "待办"}); todo != "" {
		cards = append(cards, AssistantActionCard{
			ID: "todo-1", Type: "todo", Title: "待办草稿", Detail: todo,
			TodoText: todo, Confidence: 0.7,
		})
	}
	if meetup := latestAssistantSentenceContaining(recent, []string{"见面", "约一下", "一起吃", "一起跑", "散步"}); meetup != "" {
		timeText := strings.TrimSpace(assistantTimePattern.FindString(meetup))
		if timeText == "" {
			timeText = latestAssistantPatternMatch(recent, assistantTimePattern)
		}
		location := strings.TrimSpace(assistantLocationPattern.FindString(meetup))
		if location == "" {
			location = latestAssistantPatternMatch(recent, assistantLocationPattern)
		}
		cards = append(cards, AssistantActionCard{
			ID: "meetup-1", Type: "meetup", Title: "约见草稿", Detail: meetup,
			TimeText: timeText, Location: location, Confidence: 0.76,
		})
	}
	return AssistantActionsPayload{Cards: cards}
}

func sanitizeAssistantInput(value string) string {
	value = assistantTruncate(redactAssistantContacts(strings.TrimSpace(value)), 1000)
	if assistantInjectionPattern.MatchString(value) {
		return "[已忽略不安全指令]"
	}
	return value
}

func sanitizeAssistantOutput(value string) string {
	value = assistantTruncate(redactAssistantContacts(strings.TrimSpace(value)), 500)
	if assistantInjectionPattern.MatchString(value) {
		return "[已隐藏越权内容]"
	}
	return value
}

func redactAssistantContacts(value string) string {
	value = assistantEmailPattern.ReplaceAllString(value, "[已隐藏联系方式]")
	value = assistantPhonePattern.ReplaceAllString(value, "[已隐藏联系方式]")
	return assistantContactPattern.ReplaceAllString(value, "[已隐藏联系方式]")
}

func validAssistantStarters(payload AssistantStartersPayload) bool {
	if len(payload.Starters) != 3 {
		return false
	}
	for _, starter := range payload.Starters {
		if strings.TrimSpace(starter) == "" || utf8.RuneCountInString(starter) > 160 {
			return false
		}
	}
	return true
}

func validAssistantSummary(payload AssistantSummaryPayload) bool {
	return strings.TrimSpace(payload.Summary) != "" &&
		len(payload.KeyPoints) <= 8 && len(payload.OpenQuestions) <= 8
}

func validAssistantGroupSummary(payload AssistantGroupSummaryPayload) bool {
	return strings.TrimSpace(payload.Summary) != "" && len(payload.Agreements) <= 8 &&
		len(payload.OpenQuestions) <= 8 && len(payload.NextSteps) <= 8
}

func validAssistantActionType(value string) bool {
	return value == "time" || value == "location" || value == "todo" || value == "meetup"
}

func sanitizeAssistantSummary(payload AssistantSummaryPayload) AssistantSummaryPayload {
	payload.Summary = sanitizeAssistantOutput(payload.Summary)
	payload.KeyPoints = sanitizeAssistantStrings(payload.KeyPoints)
	payload.OpenQuestions = sanitizeAssistantStrings(payload.OpenQuestions)
	return payload
}

func sanitizeAssistantGroupSummary(payload AssistantGroupSummaryPayload) AssistantGroupSummaryPayload {
	payload.Summary = sanitizeAssistantOutput(payload.Summary)
	payload.Agreements = sanitizeAssistantStrings(payload.Agreements)
	payload.OpenQuestions = sanitizeAssistantStrings(payload.OpenQuestions)
	payload.NextSteps = sanitizeAssistantStrings(payload.NextSteps)
	return payload
}

func sanitizeAssistantStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if sanitized := sanitizeAssistantOutput(value); sanitized != "" {
			result = append(result, sanitized)
		}
	}
	return result
}

func assistantRecentMessages(messages []assistantMessage, limit int) []assistantMessage {
	if len(messages) <= limit {
		return messages
	}
	return messages[len(messages)-limit:]
}

func uniqueAssistantSnippets(
	messages []assistantMessage,
	limit int,
	include func(assistantMessage) bool,
) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, limit)
	for index := len(messages) - 1; index >= 0 && len(result) < limit; index-- {
		message := messages[index]
		if !include(message) {
			continue
		}
		snippet := assistantTruncate(message.Content, 80)
		if snippet == "" {
			continue
		}
		if _, ok := seen[snippet]; ok {
			continue
		}
		seen[snippet] = struct{}{}
		result = append(result, snippet)
	}
	return result
}

func latestAssistantPatternMatch(messages []assistantMessage, pattern *regexp.Regexp) string {
	for index := len(messages) - 1; index >= 0; index-- {
		if messages[index].Content == "[已忽略不安全指令]" {
			continue
		}
		if value := strings.TrimSpace(pattern.FindString(messages[index].Content)); value != "" {
			return value
		}
	}
	return ""
}

func latestAssistantSentenceContaining(messages []assistantMessage, needles []string) string {
	for index := len(messages) - 1; index >= 0; index-- {
		if messages[index].Content == "[已忽略不安全指令]" {
			continue
		}
		if sentence := assistantSentenceContaining(messages[index].Content, needles); sentence != "" {
			return sentence
		}
	}
	return ""
}

func assistantSentenceContaining(value string, needles []string) string {
	for _, sentence := range strings.FieldsFunc(value, func(r rune) bool {
		return r == '。' || r == '！' || r == '？' || r == '\n' || r == ';' || r == '；'
	}) {
		sentence = strings.TrimSpace(sentence)
		for _, needle := range needles {
			if strings.Contains(sentence, needle) {
				return assistantTruncate(sentence, 120)
			}
		}
	}
	return ""
}

func assistantActionsConfidence(payload AssistantActionsPayload) float64 {
	if len(payload.Cards) == 0 {
		return 0.4
	}
	total := 0.0
	for _, card := range payload.Cards {
		total += card.Confidence
	}
	return total / float64(len(payload.Cards))
}

func assistantTruncate(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit-1]) + "…"
}

func deterministicAssistantMetadata(prompt ai.Prompt) ai.Metadata {
	return ai.Metadata{Source: "deterministic", PromptVersion: prompt.Version}
}

func assistantPreferencesView(preferences store.AIUserPreferences) AssistantPreferencesView {
	view := AssistantPreferencesView{Enabled: preferences.AssistantEnabled}
	if preferences.AssistantMemoryResetAt.Valid {
		resetAt := preferences.AssistantMemoryResetAt.Time
		view.MemoryResetAt = &resetAt
	}
	return view
}

func assistantArtifactView(artifact store.ChatAssistantArtifact) AssistantArtifactView {
	return AssistantArtifactView{
		ID: artifact.ID, ArtifactType: artifact.ArtifactType,
		Data:            append(json.RawMessage(nil), artifact.Payload...),
		MessageCutoffID: artifact.MessageCutoffID,
		InferenceSource: artifact.InferenceSource, InferenceModel: artifact.InferenceModel,
		Confidence: artifact.Confidence, PromptVersion: artifact.PromptVersion,
		CreatedAt: artifact.CreatedAt.Format(time.RFC3339),
	}
}

func mapAssistantContextError(err error, message string) error {
	if errors.Is(err, store.ErrNotFound) {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, message)
	}
	return httpapi.InternalError(err)
}

func invalidAssistantRequest(message string) error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, message)
}

func (s *Service) now() time.Time {
	if s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

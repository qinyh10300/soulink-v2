package questionnaire

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	questionnaireconfig "soulink-backend/configs/questionnaire"
)

const (
	defaultCatalogFile        = "thu_questionnaire_template.md"
	defaultCollectionTimezone = "Asia/Shanghai"

	labelTemplateKey         = "\u6a21\u677f\u6807\u8bc6"
	labelQuestionnaireKey    = "\u95ee\u5377\u6807\u8bc6"
	labelTemplateVersion     = "\u6a21\u677f\u7248\u672c"
	labelOpenStartTime       = "\u5f00\u653e\u5f00\u59cb\u65f6\u95f4"
	labelCollectStartTime    = "\u6536\u96c6\u5f00\u59cb\u65f6\u95f4"
	labelOpenEndTime         = "\u5f00\u653e\u7ed3\u675f\u65f6\u95f4"
	labelCollectEndTime      = "\u6536\u96c6\u7ed3\u675f\u65f6\u95f4"
	labelOpenTimezone        = "\u5f00\u653e\u65f6\u533a"
	labelCollectTimezone     = "\u6536\u96c6\u65f6\u533a"
	labelQuestionType        = "\u9898\u578b"
	labelRequired            = "\u662f\u5426\u5fc5\u7b54"
	labelRange               = "\u53d6\u503c\u8303\u56f4"
	labelScaleRange          = "\u91cf\u8868\u8303\u56f4"
	labelExtra               = "\u989d\u5916\u8bf4\u660e"
	labelLeftAnchor          = "\u5de6\u951a\u70b9"
	labelRightAnchor         = "\u53f3\u951a\u70b9"
	labelChoiceLimit         = "\u9009\u62e9\u9650\u5236"
	labelImportance          = "\u652f\u6301\u8bbe\u4e3a\u91cd\u8981"
	labelMaxLength           = "\u6700\u5927\u957f\u5ea6"
	labelPlaceholder         = "\u5360\u4f4d\u63d0\u793a"
	labelOptions             = "\u9009\u9879"
	valueYes                 = "\u662f"
	valueUnlimited           = "\u4e0d\u9650"
	valueNoFixedOptions      = "\u65e0\u56fa\u5b9a\u9009\u9879"
	questionTypeSingleChoice = "\u5355\u9009\u9898"
	questionTypeSelect       = "\u4e0b\u62c9\u9009\u62e9\u9898"
	questionTypeMultiple     = "\u591a\u9009\u9898"
	questionTypeNumber       = "\u6570\u503c\u9898"
	questionTypeRange        = "\u53cc\u7aef\u8303\u56f4\u9898"
	questionTypeScale        = "7 \u70b9\u91cf\u8868\u9898"
	questionTypeText         = "\u6587\u672c\u9898"
)

var questionHeadingPattern = regexp.MustCompile(`^([0-9]+[a-z]?)\.\s*(.+)$`)
var orderedListPattern = regexp.MustCompile(`^\d+\.\s*(.+)$`)
var numberPattern = regexp.MustCompile(`\d+`)
var keyTokenPattern = regexp.MustCompile(`[a-z0-9]+`)

func LoadCatalog(fileName string) (*Catalog, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = defaultCatalogFile
	}
	raw, err := questionnaireconfig.Files.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read questionnaire template %q: %w", fileName, err)
	}
	return ParseCatalog(string(raw))
}

func LoadDefaultCatalog() (*Catalog, error) {
	return LoadCatalog(defaultCatalogFile)
}

func ParseCatalog(markdown string) (*Catalog, error) {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")

	catalog := &Catalog{
		ID:      1,
		Version: "v1",
		byID:    make(map[string]Question),
	}

	var section string
	var current *Question
	sectionIndexByTitle := make(map[string]int)
	var collectionStartRaw string
	var collectionEndRaw string
	var collectionTimezoneRaw string
	optionMode := false

	flushCurrent := func() {
		if current == nil {
			return
		}
		current.Order = len(catalog.Questions) + 1
		catalog.Questions = append(catalog.Questions, *current)
		catalog.byID[current.ID] = *current
		if current.Required {
			catalog.RequiredCount++
		}
	}

	for _, rawLine := range lines {
		line := strings.TrimSpace(strings.TrimPrefix(rawLine, "\uFEFF"))
		if line == "" {
			optionMode = false
			continue
		}

		if strings.HasPrefix(line, "# ") {
			if current != nil {
				flushCurrent()
				current = nil
			}
			heading := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			if catalog.Title == "" {
				catalog.Title = heading
			} else {
				section = heading
			}
			optionMode = false
			continue
		}

		if strings.HasPrefix(line, "## ") {
			if current != nil {
				flushCurrent()
				current = nil
			}

			heading := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if code, title, err := parseQuestionHeading(heading); err == nil {
				current = &Question{ID: code, Section: section, Title: title, SupportsImportance: true}
			} else {
				section = heading
				if _, exists := sectionIndexByTitle[section]; !exists {
					catalog.Sections = append(catalog.Sections, SectionMeta{
						ID:    normalizeSectionID(section),
						Title: section,
					})
					sectionIndexByTitle[section] = len(catalog.Sections) - 1
				}
			}
			optionMode = false
			continue
		}

		if strings.HasPrefix(line, "### ") {
			if current != nil {
				flushCurrent()
			}

			code, title, err := parseQuestionHeading(strings.TrimSpace(strings.TrimPrefix(line, "### ")))
			if err != nil {
				return nil, err
			}
			current = &Question{ID: code, Section: section, Title: title, SupportsImportance: true}
			optionMode = false
			continue
		}

		item, isListItem := parseListItem(line)
		if current == nil {
			if !isListItem {
				continue
			}
			switch {
			case hasFieldPrefix(item, labelTemplateKey, labelQuestionnaireKey):
				catalog.Key = normalizeCatalogKey(mustFieldValue(item, labelTemplateKey, labelQuestionnaireKey))
			case hasFieldPrefix(item, labelTemplateVersion):
				catalog.Version = normalizeCatalogVersion(mustFieldValue(item, labelTemplateVersion))
			case hasFieldPrefix(item, labelOpenStartTime, labelCollectStartTime):
				collectionStartRaw = mustFieldValue(item, labelOpenStartTime, labelCollectStartTime)
			case hasFieldPrefix(item, labelOpenEndTime, labelCollectEndTime):
				collectionEndRaw = mustFieldValue(item, labelOpenEndTime, labelCollectEndTime)
			case hasFieldPrefix(item, labelOpenTimezone, labelCollectTimezone):
				collectionTimezoneRaw = mustFieldValue(item, labelOpenTimezone, labelCollectTimezone)
			}
			continue
		}

		if optionMode && isListItem {
			optionText := normalizeOptionText(item)
			switch {
			case isQuestionField(optionText):
				optionMode = false
			case looksLikeFieldLine(optionText):
				continue
			default:
				if optionText != "" && optionText != valueNoFixedOptions {
					current.Options = append(current.Options, Option{Label: optionText, Value: optionText})
				}
				continue
			}
		}

		if !isListItem {
			if current == nil && section != "" {
				if index, ok := sectionIndexByTitle[section]; ok {
					applySectionNarrative(&catalog.Sections[index], line)
				}
			}
			optionMode = false
			continue
		}

		switch {
		case hasFieldPrefix(item, labelQuestionType):
			current.Type, current.Widget = mapQuestionType(mustFieldValue(item, labelQuestionType))
			optionMode = false
		case hasFieldPrefix(item, labelRequired):
			current.Required = parseYesNo(mustFieldValue(item, labelRequired))
			optionMode = false
		case hasFieldPrefix(item, labelRange):
			min, max := parseIntRange(mustFieldValue(item, labelRange))
			current.Min = min
			current.Max = max
			optionMode = false
		case hasFieldPrefix(item, labelScaleRange):
			min, max := parseIntRange(mustFieldValue(item, labelScaleRange))
			current.Min = min
			current.Max = max
			optionMode = false
		case hasFieldPrefix(item, labelExtra):
			extra := mustFieldValue(item, labelExtra)
			if strings.Contains(extra, valueUnlimited) {
				current.AllowUnlimited = true
			}
			optionMode = false
		case hasFieldPrefix(item, labelLeftAnchor):
			current.LeftLabel = mustFieldValue(item, labelLeftAnchor)
			optionMode = false
		case hasFieldPrefix(item, labelRightAnchor):
			current.RightLabel = mustFieldValue(item, labelRightAnchor)
			optionMode = false
		case hasFieldPrefix(item, labelChoiceLimit):
			minSelections, maxSelections := parseChoiceLimit(mustFieldValue(item, labelChoiceLimit))
			current.MinSelections = minSelections
			current.MaxSelections = maxSelections
			optionMode = false
		case hasFieldPrefix(item, labelImportance):
			current.SupportsImportance = parseYesNo(mustFieldValue(item, labelImportance))
			optionMode = false
		case hasFieldPrefix(item, labelMaxLength):
			current.MaxLength = parseFirstInt(mustFieldValue(item, labelMaxLength))
			optionMode = false
		case hasFieldPrefix(item, labelPlaceholder):
			current.Placeholder = strings.Trim(mustFieldValue(item, labelPlaceholder), `"`)
			optionMode = false
		case hasFieldPrefix(item, labelOptions):
			optionValue := normalizeOptionText(mustFieldValue(item, labelOptions))
			switch {
			case optionValue == "":
				optionMode = true
			case optionValue == valueNoFixedOptions:
				optionMode = false
			default:
				current.Options = append(current.Options, Option{Label: optionValue, Value: optionValue})
				optionMode = false
			}
		default:
			if supportsExplicitOptions(current.Type) && !isQuestionField(item) && !looksLikeFieldLine(item) {
				optionValue := normalizeOptionText(item)
				if optionValue != "" && optionValue != valueNoFixedOptions {
					current.Options = append(current.Options, Option{Label: optionValue, Value: optionValue})
					optionMode = false
					continue
				}
			}
			optionMode = false
		}
	}

	if current != nil {
		flushCurrent()
	}

	catalog.TotalCount = len(catalog.Questions)
	if catalog.Title == "" || catalog.TotalCount == 0 {
		return nil, fmt.Errorf("questionnaire catalog is empty")
	}
	if catalog.Key == "" {
		catalog.Key = deriveCatalogKey(catalog.Title)
	}
	if err := applyCollectionWindow(catalog, collectionStartRaw, collectionEndRaw, collectionTimezoneRaw); err != nil {
		return nil, err
	}

	sort.SliceStable(catalog.Questions, func(i, j int) bool {
		return catalog.Questions[i].Order < catalog.Questions[j].Order
	})

	return catalog, nil
}

func parseQuestionHeading(line string) (string, string, error) {
	matches := questionHeadingPattern.FindStringSubmatch(line)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid question heading: %s", line)
	}
	return matches[1], matches[2], nil
}

func parseListItem(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "- ") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), true
	}
	matches := orderedListPattern.FindStringSubmatch(trimmed)
	if len(matches) == 2 {
		return strings.TrimSpace(matches[1]), true
	}
	return "", false
}

func hasFieldPrefix(line string, labels ...string) bool {
	_, ok := fieldValue(line, labels...)
	return ok
}

func mustFieldValue(line string, labels ...string) string {
	value, _ := fieldValue(line, labels...)
	return value
}

func fieldValue(line string, labels ...string) (string, bool) {
	for _, label := range labels {
		for _, separator := range []string{"\uff1a", ":"} {
			prefix := label + separator
			if strings.HasPrefix(line, prefix) {
				return strings.TrimSpace(strings.TrimPrefix(line, prefix)), true
			}
		}
	}
	return "", false
}

func normalizeOptionText(value string) string {
	return strings.TrimSpace(strings.Trim(value, `"`))
}

func normalizeSectionID(title string) string {
	var builder strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastHyphen = false
		case unicode.Is(unicode.Han, r):
			builder.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen && builder.Len() > 0 {
				builder.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}

func applySectionNarrative(section *SectionMeta, line string) {
	if section == nil {
		return
	}
	text := strings.TrimSpace(line)
	text = strings.TrimPrefix(text, ">")
	text = strings.TrimSpace(text)
	text = strings.Trim(text, "_")
	text = strings.Trim(text, "*")
	if text == "" {
		return
	}
	if section.Description == "" {
		section.Description = text
		return
	}
	if section.Tip == "" && strings.HasPrefix(strings.TrimSpace(line), ">") {
		section.Tip = text
	}
}

func isQuestionField(line string) bool {
	return hasFieldPrefix(line,
		labelQuestionType,
		labelRequired,
		labelRange,
		labelScaleRange,
		labelExtra,
		labelLeftAnchor,
		labelRightAnchor,
		labelChoiceLimit,
		labelImportance,
		labelMaxLength,
		labelPlaceholder,
		labelOptions,
	)
}

func mapQuestionType(raw string) (QuestionType, string) {
	switch strings.TrimSpace(raw) {
	case questionTypeSingleChoice:
		fallthrough
	case "单选":
		return QuestionTypeSingleChoice, "radio"
	case questionTypeSelect:
		return QuestionTypeSingleChoice, "select"
	case questionTypeMultiple:
		fallthrough
	case "多选":
		return QuestionTypeMultipleChoice, "checkbox"
	case questionTypeNumber:
		return QuestionTypeNumber, "number"
	case questionTypeRange:
		return QuestionTypeRange, "range"
	case questionTypeScale:
		return QuestionTypeScale, "scale"
	case questionTypeText:
		return QuestionTypeText, "textarea"
	default:
		return QuestionTypeText, "text"
	}
}

func supportsExplicitOptions(questionType QuestionType) bool {
	return questionType == QuestionTypeSingleChoice || questionType == QuestionTypeMultipleChoice
}

func parseYesNo(value string) bool {
	return strings.TrimSpace(value) == valueYes
}

func parseIntRange(value string) (*int, *int) {
	values := parseAllInts(value)
	if len(values) < 2 {
		return nil, nil
	}
	min := values[0]
	max := values[1]
	return &min, &max
}

func parseChoiceLimit(value string) (*int, *int) {
	values := parseAllInts(value)
	if len(values) == 0 {
		return nil, nil
	}
	min := values[0]
	if len(values) == 1 {
		return &min, nil
	}
	max := values[1]
	return &min, &max
}

func parseFirstInt(value string) *int {
	values := parseAllInts(value)
	if len(values) == 0 {
		return nil
	}
	first := values[0]
	return &first
}

func parseAllInts(value string) []int {
	raw := numberPattern.FindAllString(value, -1)
	result := make([]int, 0, len(raw))
	for _, item := range raw {
		parsed := 0
		for _, ch := range item {
			parsed = parsed*10 + int(ch-'0')
		}
		result = append(result, parsed)
	}
	return result
}

func looksLikeFieldLine(line string) bool {
	return strings.Contains(line, "\uff1a") || strings.Contains(line, ":")
}

func normalizeCatalogKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	tokens := keyTokenPattern.FindAllString(value, -1)
	return strings.Join(tokens, "-")
}

func normalizeCatalogVersion(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "v1"
	}
	return value
}

func deriveCatalogKey(title string) string {
	normalized := normalizeCatalogKey(title)
	if normalized != "" {
		return normalized
	}
	return "questionnaire"
}

func applyCollectionWindow(catalog *Catalog, startRaw, endRaw, timezoneRaw string) error {
	if catalog == nil {
		return nil
	}
	if startRaw == "" && endRaw == "" && strings.TrimSpace(timezoneRaw) == "" {
		return nil
	}

	timezoneName := strings.TrimSpace(timezoneRaw)
	if timezoneName == "" {
		timezoneName = defaultCollectionTimezone
	}
	location, err := time.LoadLocation(timezoneName)
	if err != nil {
		return fmt.Errorf("load questionnaire collection timezone %q: %w", timezoneName, err)
	}

	catalog.CollectionTimezone = timezoneName
	if strings.TrimSpace(startRaw) != "" {
		parsed, err := parseCatalogTime(strings.TrimSpace(startRaw), location)
		if err != nil {
			return fmt.Errorf("parse questionnaire collection start time: %w", err)
		}
		catalog.CollectionStartsAt = parsed
	}
	if strings.TrimSpace(endRaw) != "" {
		parsed, err := parseCatalogTime(strings.TrimSpace(endRaw), location)
		if err != nil {
			return fmt.Errorf("parse questionnaire collection end time: %w", err)
		}
		catalog.CollectionEndsAt = parsed
	}
	if catalog.CollectionStartsAt != nil && catalog.CollectionEndsAt != nil && catalog.CollectionEndsAt.Before(*catalog.CollectionStartsAt) {
		return fmt.Errorf("questionnaire collection end time must be after start time")
	}
	return nil
}

func parseCatalogTime(value string, location *time.Location) (*time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		result := parsed
		return &result, nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006-01-02",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			result := parsed
			return &result, nil
		}
	}

	return nil, fmt.Errorf("unsupported time format %q", value)
}

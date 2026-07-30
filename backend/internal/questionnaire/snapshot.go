package questionnaire

import (
	"encoding/json"
	"fmt"
	"sort"

	"soulink-backend/internal/store"
)

func buildAnswerSnapshotForCatalog(catalog *Catalog, records []store.QuestionnaireAnswer) (int, int, []Answer, error) {
	answerList := make([]Answer, 0, len(records))
	answeredIDs := make(map[string]struct{}, len(records))

	for _, record := range records {
		question, ok := catalog.QuestionByID(record.QuestionID)
		if !ok {
			continue
		}

		var value any
		if err := json.Unmarshal(record.ValueJSON, &value); err != nil {
			return 0, 0, nil, fmt.Errorf("decode questionnaire answer %s: %w", record.QuestionID, err)
		}

		answeredIDs[record.QuestionID] = struct{}{}
		answerList = append(answerList, Answer{
			QuestionID: question.ID,
			Value:      value,
			Importance: record.IsImportant,
		})
	}

	requiredAnsweredCount := 0
	for _, question := range catalog.Questions {
		if question.Required {
			if _, ok := answeredIDs[question.ID]; ok {
				requiredAnsweredCount++
			}
		}
	}

	sort.SliceStable(answerList, func(i, j int) bool {
		left, _ := catalog.QuestionByID(answerList[i].QuestionID)
		right, _ := catalog.QuestionByID(answerList[j].QuestionID)
		if left.Order == right.Order {
			return answerList[i].QuestionID < answerList[j].QuestionID
		}
		return left.Order < right.Order
	})

	return len(answeredIDs), requiredAnsweredCount, answerList, nil
}

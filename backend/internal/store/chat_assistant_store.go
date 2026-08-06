package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type ChatAssistantArtifact struct {
	ID              int64
	UserID          int64
	ContextType     string
	ContextID       int64
	ArtifactType    string
	Payload         json.RawMessage
	MessageCutoffID int64
	InferenceSource string
	InferenceModel  string
	Confidence      float64
	PromptVersion   string
	CreatedAt       time.Time
	DeletedAt       sql.NullTime
}

type CreateChatAssistantArtifactParams struct {
	UserID          int64
	ContextType     string
	ContextID       int64
	ArtifactType    string
	Payload         json.RawMessage
	MessageCutoffID int64
	InferenceSource string
	InferenceModel  string
	Confidence      float64
	PromptVersion   string
	Now             time.Time
}

func (s *Store) CreateChatAssistantArtifact(
	ctx context.Context,
	tx *sql.Tx,
	params CreateChatAssistantArtifactParams,
) (ChatAssistantArtifact, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO chat_assistant_artifacts (
			user_id, context_type, context_id, artifact_type, payload,
			message_cutoff_id, inference_source, inference_model,
			confidence, prompt_version, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, context_type, context_id, artifact_type, payload,
		          message_cutoff_id, inference_source, inference_model,
		          confidence, prompt_version, created_at, deleted_at
	`, params.UserID, params.ContextType, params.ContextID, params.ArtifactType,
		[]byte(params.Payload), params.MessageCutoffID, params.InferenceSource,
		params.InferenceModel, params.Confidence, params.PromptVersion, params.Now)
	artifact, err := scanChatAssistantArtifact(row)
	if err != nil {
		return ChatAssistantArtifact{}, fmt.Errorf("create chat assistant artifact: %w", err)
	}
	return artifact, nil
}

func (s *Store) ListLatestChatAssistantArtifacts(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	contextType string,
	contextID int64,
) ([]ChatAssistantArtifact, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT id, user_id, context_type, context_id, artifact_type, payload,
		       message_cutoff_id, inference_source, inference_model,
		       confidence, prompt_version, created_at, deleted_at
		FROM (
			SELECT DISTINCT ON (artifact_type)
			       id, user_id, context_type, context_id, artifact_type, payload,
			       message_cutoff_id, inference_source, inference_model,
			       confidence, prompt_version, created_at, deleted_at
			FROM chat_assistant_artifacts
			WHERE user_id = $1 AND context_type = $2 AND context_id = $3
			  AND deleted_at IS NULL
			ORDER BY artifact_type, created_at DESC, id DESC
		) latest
		ORDER BY created_at DESC, id DESC
	`, userID, contextType, contextID)
	if err != nil {
		return nil, fmt.Errorf("list latest chat assistant artifacts: %w", err)
	}
	defer rows.Close()

	artifacts := make([]ChatAssistantArtifact, 0)
	for rows.Next() {
		artifact, err := scanChatAssistantArtifact(rows)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat assistant artifacts: %w", err)
	}
	return artifacts, nil
}

func (s *Store) SoftDeleteChatAssistantArtifact(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	contextType string,
	contextID, artifactID int64,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE chat_assistant_artifacts
		SET deleted_at = $5
		WHERE id = $1 AND user_id = $2 AND context_type = $3 AND context_id = $4
		  AND deleted_at IS NULL
	`, artifactID, userID, contextType, contextID, now)
	if err != nil {
		return fmt.Errorf("delete chat assistant artifact: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) SetAssistantEnabled(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	enabled bool,
	now time.Time,
) (AIUserPreferences, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO ai_user_preferences (
			user_id, assistant_enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET assistant_enabled = EXCLUDED.assistant_enabled,
		    updated_at = EXCLUDED.updated_at
		RETURNING user_id, recommendation_learning_enabled, assistant_enabled,
		          learning_reset_at, assistant_memory_reset_at, created_at, updated_at
	`, userID, enabled, now)
	preferences, err := scanAIUserPreferences(row)
	if err != nil {
		return AIUserPreferences{}, fmt.Errorf("set assistant enabled: %w", err)
	}
	return preferences, nil
}

func (s *Store) ResetAssistantMemory(ctx context.Context, tx *sql.Tx, userID int64, now time.Time) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO ai_user_preferences (
			user_id, assistant_memory_reset_at, created_at, updated_at
		) VALUES ($1, $2, $2, $2)
		ON CONFLICT (user_id) DO UPDATE
		SET assistant_memory_reset_at = EXCLUDED.assistant_memory_reset_at,
		    updated_at = EXCLUDED.updated_at
	`, userID, now); err != nil {
		return fmt.Errorf("reset assistant memory preference: %w", err)
	}
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE chat_assistant_artifacts
		SET deleted_at = $2
		WHERE user_id = $1 AND deleted_at IS NULL
	`, userID, now); err != nil {
		return fmt.Errorf("delete assistant memory artifacts: %w", err)
	}
	return nil
}

func scanChatAssistantArtifact(scanner rowScanner) (ChatAssistantArtifact, error) {
	var artifact ChatAssistantArtifact
	var payload []byte
	if err := scanner.Scan(
		&artifact.ID, &artifact.UserID, &artifact.ContextType, &artifact.ContextID,
		&artifact.ArtifactType, &payload, &artifact.MessageCutoffID,
		&artifact.InferenceSource, &artifact.InferenceModel, &artifact.Confidence,
		&artifact.PromptVersion, &artifact.CreatedAt, &artifact.DeletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return ChatAssistantArtifact{}, ErrNotFound
		}
		return ChatAssistantArtifact{}, fmt.Errorf("scan chat assistant artifact: %w", err)
	}
	artifact.Payload = append(json.RawMessage(nil), payload...)
	return artifact, nil
}

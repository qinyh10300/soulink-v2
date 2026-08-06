package operations

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"soulink-backend/internal/ai"
)

type generatorStub struct {
	err error
}

func (stub generatorStub) Generate(
	context.Context,
	ai.Prompt,
	any,
	any,
) (ai.Metadata, error) {
	return ai.Metadata{Source: "model"}, stub.err
}

func TestInstrumentGeneratorRecordsFailures(t *testing.T) {
	collector := NewCollector(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	wantErr := errors.New("provider unavailable")
	generator := InstrumentGenerator(generatorStub{err: wantErr}, collector)
	if _, err := generator.Generate(t.Context(), ai.Prompt{}, nil, &struct{}{}); !errors.Is(err, wantErr) {
		t.Fatalf("Generate() error = %v", err)
	}
	snapshot := collector.Snapshot(time.Now())
	if len(snapshot.Tasks) != 1 || snapshot.Tasks[0].Name != "ai.generate" || snapshot.Tasks[0].Failures != 1 {
		t.Fatalf("tasks = %+v", snapshot.Tasks)
	}
}

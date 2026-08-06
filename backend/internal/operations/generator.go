package operations

import (
	"context"
	"time"

	"soulink-backend/internal/ai"
)

type instrumentedGenerator struct {
	delegate  ai.Generator
	collector *Collector
}

func InstrumentGenerator(delegate ai.Generator, collector *Collector) ai.Generator {
	if delegate == nil {
		return nil
	}
	return &instrumentedGenerator{delegate: delegate, collector: collector}
}

func (g *instrumentedGenerator) Generate(
	ctx context.Context,
	prompt ai.Prompt,
	input any,
	output any,
) (ai.Metadata, error) {
	startedAt := time.Now()
	metadata, err := g.delegate.Generate(ctx, prompt, input, output)
	g.collector.ObserveTask(ctx, "ai.generate", time.Since(startedAt), err)
	return metadata, err
}

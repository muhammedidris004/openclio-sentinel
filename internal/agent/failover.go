package agent

import (
	"context"
	"fmt"
	"log/slog"
)

// FailoverProvider tries multiple providers in order, falling through on error.
type FailoverProvider struct {
	providers []Provider
	logger    *slog.Logger
}

// NewFailoverProvider creates a provider that tries primary first, then fallbacks.
func NewFailoverProvider(primary Provider, fallbacks []Provider, logger *slog.Logger) *FailoverProvider {
	all := append([]Provider{primary}, fallbacks...)
	return &FailoverProvider{providers: all, logger: logger}
}

func (fp *FailoverProvider) Name() string {
	if len(fp.providers) > 0 {
		return "failover(" + fp.providers[0].Name() + ")"
	}
	return "failover(empty)"
}

func (fp *FailoverProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	var lastErr error

	for i, p := range fp.providers {
		if i == 0 {
			if fp.logger != nil {
				fp.logger.Debug("trying primary provider", "provider", p.Name())
			}
		} else {
			if fp.logger != nil {
				fp.logger.Warn("primary failed, trying fallback",
					"fallback", p.Name(),
					"error", lastErr.Error(),
				)
			}
		}

		resp, err := p.Chat(ctx, req)
		if err == nil {
			if i > 0 && fp.logger != nil {
				fp.logger.Info("failover succeeded", "provider", p.Name())
			}
			return resp, nil
		}

		lastErr = err

		// If context is done, no point trying further
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Don't retry on permanent non-retryable errors from primary
		if i == 0 && !IsRetryable(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

func (fp *FailoverProvider) Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	var lastErr error

	for i, p := range fp.providers {
		streamer, ok := p.(Streamer)
		if !ok {
			// Degrade gracefully by synthesizing a single buffered chunk for non-stream providers.
			resp, err := p.Chat(ctx, req)
			if err == nil {
				ch := make(chan StreamChunk, 2)
				ch <- StreamChunk{Text: resp.Content}
				ch <- StreamChunk{ToolCalls: resp.ToolCalls, Done: true}
				close(ch)
				return ch, nil
			}
			lastErr = err
		} else {
			ch, err := streamer.Stream(ctx, req)
			if err == nil {
				if i > 0 && fp.logger != nil {
					fp.logger.Info("failover stream succeeded", "provider", p.Name())
				}
				return ch, nil
			}
			lastErr = err
		}

		if i == 0 {
			if fp.logger != nil {
				fp.logger.Debug("trying primary stream provider", "provider", p.Name())
			}
		} else if fp.logger != nil && lastErr != nil {
			fp.logger.Warn("primary stream failed, trying fallback",
				"fallback", p.Name(),
				"error", lastErr.Error(),
			)
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if i == 0 && lastErr != nil && !IsRetryable(lastErr) {
			return nil, lastErr
		}
	}

	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

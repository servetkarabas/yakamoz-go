package ai

import (
	"context"
	"fmt"
)

type TranslateRequest struct {
	Text           string
	SourceLanguage string
	TargetLanguage string
}

type TranslateResult struct {
	Text string
}

type Translator interface {
	Translate(context.Context, TranslateRequest) (TranslateResult, error)
}

type Verdict string

const (
	VerdictAllow  Verdict = "allow"
	VerdictReview Verdict = "review"
	VerdictBlock  Verdict = "block"
)

type Moderator interface {
	Classify(context.Context, string, string) (Verdict, error)
}

type Stub struct{}

func (Stub) Translate(ctx context.Context, req TranslateRequest) (TranslateResult, error) {
	select {
	case <-ctx.Done():
		return TranslateResult{}, ctx.Err()
	default:
	}
	return TranslateResult{Text: fmt.Sprintf("[%s] %s", req.TargetLanguage, req.Text)}, nil
}

type StubModerator struct{}

func (StubModerator) Classify(ctx context.Context, _ string, _ string) (Verdict, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return VerdictAllow, nil
	}
}

package review

import (
	"context"

	npool "github.com/NpoolPlatform/message/npool/review/mw/v2"
)

type Handler struct {
	AppID       *string
	UserID      *string
	LangID      *string
	TargetAppID *string
	ReviewID    *string
	State       *npool.ReviewState
	Message     *string
}

func NewHandler(ctx context.Context, options ...func(context.Context, *Handler) error) (*Handler, error) {
	handler := &Handler{}
	for _, opt := range options {
		if err := opt(ctx, handler); err != nil {
			return nil, err
		}
	}
	return handler, nil
}

func WithAppID(appID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		h.AppID = appID
		return nil
	}
}

func WithTargetAppID(appID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		h.TargetAppID = appID
		return nil
	}
}

func WithUserID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		h.UserID = id
		return nil
	}
}

func WithReviewID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		h.ReviewID = id
		return nil
	}
}

func WithLangID(langID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		h.LangID = langID
		return nil
	}
}

func WithState(state *npool.ReviewState, message *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		h.State = state
		return nil
	}
}

func WithMessage(message *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if message == nil {
			return nil
		}
		h.Message = message
		return nil
	}
}
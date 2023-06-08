package withdraw

import (
	"context"
	"fmt"

	appcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/app"
	npool "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	constant "github.com/NpoolPlatform/review-gateway/pkg/const"
	"github.com/google/uuid"
)

type Handler struct {
	AppID       *string
	TargetAppID *string
	UserID      *string
	ReviewID    *uuid.UUID
	LangID      *string
	Domain      *string
	State       *npool.ReviewState
	Message     *string
	Offset      int32
	Limit       int32
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
		_, err := uuid.Parse(*appID)
		if err != nil {
			return err
		}
		exist, err := appcli.ExistApp(ctx, *appID)
		if err != nil {
			return err
		}
		if !exist {
			return fmt.Errorf("invalid app")
		}

		h.AppID = appID
		return nil
	}
}

func WithTargetAppID(appID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_, err := uuid.Parse(*appID)
		if err != nil {
			return err
		}
		exist, err := appcli.ExistApp(ctx, *appID)
		if err != nil {
			return err
		}
		if !exist {
			return fmt.Errorf("invalid app")
		}
		h.TargetAppID = appID
		return nil
	}
}

func WithUserID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_, err := uuid.Parse(*id)
		if err != nil {
			return err
		}
		// TODO: Check UserID
		h.UserID = id
		return nil
	}
}

func WithReviewID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_id, err := uuid.Parse(*id)
		if err != nil {
			return err
		}
		h.ReviewID = &_id
		return nil
	}
}

func WithState(state *npool.ReviewState, message *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if state == nil {
			return nil
		}
		switch *state {
		case npool.ReviewState_Rejected:
		case npool.ReviewState_Approved:
		default:
			return fmt.Errorf("invalid review state")
		}
		if *state == npool.ReviewState_Rejected && message == nil {
			return fmt.Errorf("message is empty")
		}
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

func WithOffset(offset int32) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		h.Offset = offset
		return nil
	}
}

func WithLimit(limit int32) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if limit == 0 {
			limit = constant.DefaultRowLimit
		}
		h.Limit = limit
		return nil
	}
}

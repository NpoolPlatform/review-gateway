package kyc

import (
	"context"
	"fmt"

	appcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/app"
	npool "github.com/NpoolPlatform/message/npool/review/mw/v2"
	constant "github.com/NpoolPlatform/review-gateway/pkg/const"
	"github.com/google/uuid"
)

type Handler struct {
	ID          *uuid.UUID
	AppID       *string
	TargetAppID *string
	UserID      *uuid.UUID
	ReviewID    *uuid.UUID
	LangID      *uuid.UUID
	Domain      *string
	ObjectID    *uuid.UUID
	ObjectIDs   []*uuid.UUID
	Trigger     *npool.ReviewTriggerType
	ObjectType  *npool.ReviewObjectType
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

func WithID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_id, err := uuid.Parse(*id)
		if err != nil {
			return err
		}
		h.ID = &_id
		return nil
	}
}

func WithAppID(appID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if appID == nil {
			return nil
		}
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
		err := WithAppID(appID)
		if err != nil {
			return fmt.Errorf("invalid target app id")
		}
		h.TargetAppID = appID
		return nil
	}
}
func WithUserID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_id, err := uuid.Parse(*id)
		if err != nil {
			return err
		}
		h.UserID = &_id
		return nil
	}
}

func WithReviewID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if id == nil {
			return nil
		}
		_id, err := uuid.Parse(*id)
		if err != nil {
			return err
		}
		h.ReviewID = &_id
		return nil
	}
}

func WithLangID(langID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if langID == nil {
			return nil
		}
		_id, err := uuid.Parse(*langID)
		if err != nil {
			return err
		}
		h.LangID = &_id
		return nil
	}
}

func WithDomain(domain *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if domain == nil {
			return fmt.Errorf("invalid domain")
		}
		h.Domain = domain
		return nil
	}
}

func WithTrigger(trigger *npool.ReviewTriggerType) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if trigger == nil {
			return nil
		}
		switch *trigger {
		case npool.ReviewTriggerType_InsufficientFunds:
		case npool.ReviewTriggerType_InsufficientGas:
		case npool.ReviewTriggerType_InsufficientFundsGas:
		case npool.ReviewTriggerType_LargeAmount:
		case npool.ReviewTriggerType_AutoReviewed:
		default:
			return fmt.Errorf("invalid trigger type")
		}

		h.Trigger = trigger
		return nil
	}
}

func WithObjectType(_type *npool.ReviewObjectType) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if _type == nil {
			return nil
		}
		switch *_type {
		case npool.ReviewObjectType_ObjectKyc:
		case npool.ReviewObjectType_ObjectWithdrawal:
		default:
			return fmt.Errorf("invalid object type")
		}
		h.ObjectType = _type
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

package kyc

import (
	"context"
	"fmt"

	kycmwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/kyc"
	appusermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	kycmwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/kyc"
	appusermwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	reviewmwpb "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

type queryHandler struct {
	*Handler
	kycs      []*kycmwpb.Kyc
	userMap   map[string]*appusermwpb.User
	reviewMap map[string]*reviewmwpb.Review
	infos     []*npool.KycReview
}

func (h *queryHandler) getReviews(ctx context.Context) error {
	ids := []string{}
	for _, kyc := range h.kycs {
		ids = append(ids, kyc.ReviewID)
	}

	infos, _, err := reviewmwcli.GetReviews(ctx, &reviewmwpb.Conds{
		ObjectType: &basetypes.Uint32Val{Op: cruder.EQ, Value: uint32(reviewtypes.ReviewObjectType_ObjectKyc)},
		EntIDs:     &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
	}, 0, int32(len(ids)))
	if err != nil {
		return err
	}

	for _, info := range infos {
		h.reviewMap[info.EntID] = info
	}
	return nil
}

func (h *queryHandler) getUsers(ctx context.Context) error {
	ids := []string{}
	for _, kyc := range h.kycs {
		ids = append(ids, kyc.UserID)
	}

	infos, _, err := appusermwcli.GetUsers(ctx, &appusermwpb.Conds{
		AppID: &basetypes.StringVal{Op: cruder.EQ, Value: *h.TargetAppID},
		IDs:   &basetypes.StringSliceVal{Op: cruder.IN, Value: ids},
	}, 0, int32(len(ids)))
	if err != nil {
		return err
	}

	for _, info := range infos {
		h.userMap[info.ID] = info
	}
	return nil
}

func (h *queryHandler) formalize() {
	for _, kyc := range h.kycs {
		rv, ok := h.reviewMap[kyc.ReviewID]
		if !ok {
			continue
		}
		user, ok := h.userMap[kyc.UserID]
		if !ok {
			continue
		}

		h.infos = append(h.infos, &npool.KycReview{
			UserID:       user.ID,
			EmailAddress: user.EmailAddress,
			PhoneNO:      user.PhoneNO,
			Username:     user.Username,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			KycID:        kyc.ID,
			DocumentType: kyc.DocumentType,
			IDNumber:     kyc.IDNumber,
			FrontImg:     kyc.FrontImg,
			BackImg:      kyc.BackImg,
			SelfieImg:    kyc.SelfieImg,
			EntityType:   kyc.EntityType,
			KycState:     kyc.State,
			ReviewID:     rv.EntID,
			ObjectType:   rv.ObjectType,
			Domain:       rv.Domain,
			Reviewer:     rv.ReviewerID,
			ReviewState:  rv.State,
			Message:      rv.Message,
			CreatedAt:    rv.CreatedAt,
			UpdatedAt:    rv.UpdatedAt,
		})
	}
}

func (h *Handler) GetKycReviews(ctx context.Context) ([]*npool.KycReview, uint32, error) {
	kycs, total, err := kycmwcli.GetKycs(ctx, &kycmwpb.Conds{
		AppID: &basetypes.StringVal{Op: cruder.EQ, Value: *h.TargetAppID},
	}, h.Offset, h.Limit)
	if err != nil {
		return nil, 0, err
	}
	if len(kycs) == 0 {
		return nil, 0, nil
	}

	handler := &queryHandler{
		Handler:   h,
		kycs:      kycs,
		userMap:   map[string]*appusermwpb.User{},
		reviewMap: map[string]*reviewmwpb.Review{},
	}

	if err := handler.getReviews(ctx); err != nil {
		return nil, 0, err
	}
	if err := handler.getUsers(ctx); err != nil {
		return nil, 0, err
	}

	handler.formalize()
	return handler.infos, total, nil
}

func (h *Handler) GetKycReview(ctx context.Context) (*npool.KycReview, error) {
	if h.KycID == nil {
		return nil, fmt.Errorf("invalid kycid")
	}

	kyc, err := kycmwcli.GetKyc(ctx, *h.KycID)
	if err != nil {
		return nil, err
	}
	if kyc == nil {
		return nil, nil
	}

	handler := &queryHandler{
		Handler:   h,
		kycs:      []*kycmwpb.Kyc{kyc},
		reviewMap: map[string]*reviewmwpb.Review{},
	}

	if err := handler.getReviews(ctx); err != nil {
		return nil, err
	}
	if err := handler.getUsers(ctx); err != nil {
		return nil, err
	}

	handler.formalize()
	if len(handler.infos) == 0 {
		return nil, nil
	}
	if len(handler.infos) > 1 {
		return nil, fmt.Errorf("too many record")
	}
	return handler.infos[0], nil
}

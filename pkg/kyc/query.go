package kyc

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	appusergateway "github.com/NpoolPlatform/appuser-gateway/pkg/message/const"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"

	kycmwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/kyc"
	usermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	kycmwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/kyc"
	usermwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	reviewpb "github.com/NpoolPlatform/message/npool/review/mw/v2"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

func (h *Handler) GetKycReviews(ctx context.Context) ([]*npool.KycReview, uint32, error) {
	kycs, total, err := kycmwcli.GetKycs(ctx, &kycmwpb.Conds{
		AppID: &basetypes.StringVal{
			Op:    cruder.EQ,
			Value: *h.AppID,
		},
	}, h.Offset, h.Limit)
	if err != nil {
		return nil, 0, err
	}
	if len(kycs) == 0 {
		return nil, 0, nil
	}

	ids := []string{}
	for _, k := range kycs {
		ids = append(ids, k.ID)
	}

	rvs, err := reviewmwcli.GetObjectReviews(
		ctx,
		*h.AppID,
		appusergateway.ServiceName,
		ids,
		reviewpb.ReviewObjectType_ObjectKyc,
	)
	if err != nil {
		return nil, 0, err
	}

	rvMap := map[string]*reviewpb.Review{}
	for _, rv := range rvs {
		rvMap[rv.ID] = rv
	}

	uids := []string{}
	for _, w := range kycs {
		uids = append(uids, w.UserID)
	}

	users, _, err := usermwcli.GetUsers(ctx, &usermwpb.Conds{
		IDs: &basetypes.StringSliceVal{Op: cruder.IN, Value: uids},
	}, 0, int32(len(uids)))
	if err != nil {
		return nil, 0, err
	}

	userMap := map[string]*usermwpb.User{}
	for _, user := range users {
		userMap[user.ID] = user
	}

	infos := []*npool.KycReview{}
	for _, kyc := range kycs {
		rv := &reviewpb.Review{}

		rvM, ok := rvMap[kyc.ReviewID]
		if ok {
			rv = rvM

			switch rv.State {
			case reviewpb.ReviewState_Approved:
			case reviewpb.ReviewState_Rejected:
			case reviewpb.ReviewState_Wait:
			default:
				logger.Sugar().Warnw("GetKycReviews", "State", rv.State)
			}
		}

		user := &usermwpb.User{}
		userM, ok := userMap[kyc.UserID]
		if ok {
			user = userM
		}

		infos = append(infos, &npool.KycReview{
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
			ReviewID:     rv.ID,
			ObjectType:   rv.ObjectType,
			Domain:       rv.Domain,
			Reviewer:     rvM.ReviewerID,
			ReviewState:  rv.State,
			KycState:     kyc.State,
			Message:      rv.Message,
			CreatedAt:    rv.CreatedAt,
			UpdatedAt:    rv.UpdatedAt,
		})
	}

	return infos, total, nil
}

// nolint
func (h *Handler) GetKycReview(ctx context.Context) (*npool.KycReview, error) {
	if h.ReviewID == nil {
		return nil, fmt.Errorf("invalid review id")
	}

	rv, err := reviewmwcli.GetReview(ctx, h.ReviewID.String())
	if err != nil {
		return nil, err
	}

	switch rv.ObjectType {
	case reviewpb.ReviewObjectType_ObjectKyc:
	default:
		return nil, fmt.Errorf("invalid object type")
	}

	kyc, err := kycmwcli.GetKyc(ctx, rv.ObjectID)
	if err != nil {
		return nil, err
	}
	if kyc == nil {
		return nil, fmt.Errorf("invalid kyc")
	}

	user, err := usermwcli.GetUser(ctx, kyc.AppID, kyc.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid user")
	}

	switch rv.State {
	case reviewpb.ReviewState_Approved:
	case reviewpb.ReviewState_Rejected:
	case reviewpb.ReviewState_Wait:
	default:
		return nil, fmt.Errorf("invalid state")
	}

	return &npool.KycReview{
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
		ReviewID:     rv.ID,
		ObjectType:   rv.ObjectType,
		Domain:       rv.Domain,
		Reviewer:     rv.ReviewerID,
		ReviewState:  rv.State,
		KycState:     kyc.State,
		Message:      rv.Message,
		CreatedAt:    rv.CreatedAt,
		UpdatedAt:    rv.UpdatedAt,
	}, nil
}

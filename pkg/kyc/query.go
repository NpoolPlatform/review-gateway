package kyc

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	appusergateway "github.com/NpoolPlatform/appuser-gateway/pkg/message/const"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"

	kyccli "github.com/NpoolPlatform/appuser-manager/pkg/client/kyc"
	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	commonpb "github.com/NpoolPlatform/message/npool"
	kycmgrpb "github.com/NpoolPlatform/message/npool/appuser/mgr/v2/kyc"
	userpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	reviewpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"
	reviewcli "github.com/NpoolPlatform/review-manager/pkg/client/review"
	reviewmwcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

// nolint
func GetkycReviews(ctx context.Context, appID string, offset, limit int32) ([]*npool.KycReview, uint32, error) {
	conds := &kycmgrpb.Conds{
		AppID: &commonpb.StringVal{
			Op:    cruder.EQ,
			Value: appID,
		},
	}
	kycs, total, err := kyccli.GetKycs(ctx, conds, offset, limit)
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
		appID,
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

	users, _, err := usercli.GetManyUsers(ctx, uids)
	if err != nil {
		return nil, 0, err
	}

	userMap := map[string]*userpb.User{}
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

		user := &userpb.User{}
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
			Reviewer:     "TODO: to be filled",
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
func GetKycReview(ctx context.Context, reviewID string) (*npool.KycReview, error) {
	rv, err := reviewcli.GetReview(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	switch rv.ObjectType {
	case reviewpb.ReviewObjectType_ObjectKyc:
	default:
		return nil, fmt.Errorf("invalid object type")
	}

	kyc, err := kyccli.GetKyc(ctx, rv.ObjectID)
	if err != nil {
		return nil, err
	}
	if kyc == nil {
		return nil, fmt.Errorf("invalid kyc")
	}

	user, err := usercli.GetUser(ctx, kyc.AppID, kyc.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid user")
	}

	// TODO: we need fill reviewer name, but we miss appid in reviews table

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
		Reviewer:     "TODO: to be filled",
		ReviewState:  rv.State,
		KycState:     kyc.State,
		Message:      rv.Message,
		CreatedAt:    rv.CreatedAt,
		UpdatedAt:    rv.UpdatedAt,
	}, nil
}

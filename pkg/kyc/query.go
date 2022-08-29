package kyc

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"

	appusergateway "github.com/NpoolPlatform/appuser-gateway/pkg/message/const"
	kycmgrconst "github.com/NpoolPlatform/kyc-management/pkg/message/const"
	reviewpb "github.com/NpoolPlatform/message/npool/review-service"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"

	kyccli "github.com/NpoolPlatform/appuser-manager/pkg/client/kyc"
	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	apisconstant "github.com/NpoolPlatform/cloud-hashing-apis/pkg/const"
	cruder "github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	commonpb "github.com/NpoolPlatform/message/npool"
	kycmgrpb "github.com/NpoolPlatform/message/npool/appuser/mgr/v2/kyc"
	userpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"
	reviewcli "github.com/NpoolPlatform/review-service/pkg/client"
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

	rvs := []*reviewpb.Review{}

	rvs, err = reviewcli.GetDomainReviews(ctx, appID, appusergateway.ServiceName,
		reviewmgrpb.ReviewObjectType_ObjectKyc.String())
	if err != nil {
		return nil, 0, err
	}
	rvs1, err := reviewcli.GetDomainReviews(ctx, appID, kycmgrconst.ServiceName,
		apisconstant.ReviewObjectKyc)
	if err != nil {
		return nil, 0, err
	}

	rvs = append(rvs, rvs1...)

	rvMap := map[string]*reviewpb.Review{}
	for _, rv := range rvs {
		rvMap[rv.ObjectID] = rv
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
		rv := &reviewpb.Review{
			ID:       "",
			ObjectID: "",
			Domain:   "",
			Message:  "",
			CreateAt: 0,
		}

		state := reviewmgrpb.ReviewState_DefaultReviewState
		trigger := reviewmgrpb.ReviewTriggerType_DefaultTriggerType

		rvM, ok := rvMap[kyc.ID]
		if ok {
			rv = rvM

			switch rv.State {
			case "approved":
				fallthrough // nolint
			case reviewmgrpb.ReviewState_Approved.String():
				state = reviewmgrpb.ReviewState_Approved
			case "rejected":
				fallthrough // nolint
			case reviewmgrpb.ReviewState_Rejected.String():
				state = reviewmgrpb.ReviewState_Rejected
			case "wait":
				fallthrough // nolint
			case reviewmgrpb.ReviewState_Wait.String():
				state = reviewmgrpb.ReviewState_Wait
			default:
				logger.Sugar().Warnw("GetKycReviews", "State", rv.State)
				continue
			}

			switch rv.Trigger {
			case "large amount":
				fallthrough // nolint
			case reviewmgrpb.ReviewTriggerType_LargeAmount.String():
				trigger = reviewmgrpb.ReviewTriggerType_LargeAmount
			case "insufficient":
				fallthrough // nolint
			case reviewmgrpb.ReviewTriggerType_InsufficientFunds.String():
				trigger = reviewmgrpb.ReviewTriggerType_InsufficientFunds
			case "auto review":
				fallthrough // nolint
			case reviewmgrpb.ReviewTriggerType_AutoReviewed.String():
				trigger = reviewmgrpb.ReviewTriggerType_AutoReviewed
			case reviewmgrpb.ReviewTriggerType_InsufficientGas.String():
				trigger = reviewmgrpb.ReviewTriggerType_InsufficientGas
			default:
				logger.Sugar().Warnw("GetKycReviews", "Trigger", rv.Trigger)
				continue
			}
		}

		user := &userpb.User{
			ID:           "",
			EmailAddress: "",
			PhoneNO:      "",
		}
		userM, ok := userMap[kyc.UserID]
		if ok {
			user = userM
		}

		infos = append(infos, &npool.KycReview{
			UserID:       user.ID,
			EmailAddress: user.EmailAddress,
			PhoneNO:      user.PhoneNO,
			KycID:        kyc.ID,
			DocumentType: kyc.DocumentType,
			IDNumber:     kyc.IDNumber,
			FrontImg:     kyc.FrontImg,
			BackImg:      kyc.BackImg,
			SelfieImg:    kyc.SelfieImg,
			EntityType:   kyc.EntityType,
			ReviewID:     rv.ID,
			ObjectType:   rv.ObjectID,
			Domain:       rv.Domain,
			Reviewer:     "TODO: to be filled",
			State:        state,
			Trigger:      trigger,
			Message:      rv.Message,
			CreatedAt:    rv.CreateAt,
			UpdatedAt:    rv.CreateAt,
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
	case "kyc":
	case reviewmgrpb.ReviewObjectType_ObjectKyc.String():
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
	state := reviewmgrpb.ReviewState_Wait

	switch rv.State {
	case "approved":
		fallthrough // nolint
	case reviewmgrpb.ReviewState_Approved.String():
		state = reviewmgrpb.ReviewState_Approved
	case "rejected":
		fallthrough // nolint
	case reviewmgrpb.ReviewState_Rejected.String():
		state = reviewmgrpb.ReviewState_Rejected
	case "wait":
		fallthrough // nolint
	case reviewmgrpb.ReviewState_Wait.String():
		state = reviewmgrpb.ReviewState_Wait
	default:
		return nil, fmt.Errorf("invalid state")
	}

	trigger := reviewmgrpb.ReviewTriggerType_InsufficientFunds

	switch rv.Trigger {
	case "large amount":
		fallthrough // nolint
	case reviewmgrpb.ReviewTriggerType_LargeAmount.String():
		trigger = reviewmgrpb.ReviewTriggerType_LargeAmount
	case "insufficient":
		fallthrough // nolint
	case reviewmgrpb.ReviewTriggerType_InsufficientFunds.String():
		trigger = reviewmgrpb.ReviewTriggerType_InsufficientFunds
	case "auto review":
		fallthrough // nolint
	case reviewmgrpb.ReviewTriggerType_AutoReviewed.String():
		trigger = reviewmgrpb.ReviewTriggerType_AutoReviewed
	case reviewmgrpb.ReviewTriggerType_InsufficientGas.String():
		trigger = reviewmgrpb.ReviewTriggerType_InsufficientGas
	default:
		return nil, fmt.Errorf("invalid trigger")
	}

	return &npool.KycReview{
		UserID:       user.ID,
		EmailAddress: user.EmailAddress,
		PhoneNO:      user.PhoneNO,
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
		State:        state,
		Trigger:      trigger,
		Message:      rv.Message,
		CreatedAt:    rv.CreateAt,
		UpdatedAt:    rv.CreateAt,
	}, nil
}

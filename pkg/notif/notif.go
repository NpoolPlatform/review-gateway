package notif

import (
	"context"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	commonpb "github.com/NpoolPlatform/message/npool"
	channelpb "github.com/NpoolPlatform/message/npool/notif/mgr/v1/channel"
	notifmgrpb "github.com/NpoolPlatform/message/npool/notif/mgr/v1/notif"
	thirdmgrpb "github.com/NpoolPlatform/message/npool/third/mgr/v1/template/notif"
	notifcli "github.com/NpoolPlatform/notif-middleware/pkg/client/notif"
	thirdcli "github.com/NpoolPlatform/third-middleware/pkg/client/template/notif"
	thirdpkg "github.com/NpoolPlatform/third-middleware/pkg/template/notif"
)

func CreateNotif(
	ctx context.Context,
	appID, userID string,
	userName,
	amount,
	coinUnit,
	address *string,
	eventType notifmgrpb.EventType,
) {
	offset := uint32(0)
	limit := uint32(1000) //nolint
	for {
		templateInfos, _, err := thirdcli.GetNotifTemplates(ctx, &thirdmgrpb.Conds{
			AppID: &commonpb.StringVal{
				Op:    cruder.EQ,
				Value: appID,
			},
			UsedFor: &commonpb.Uint32Val{
				Op:    cruder.EQ,
				Value: uint32(eventType.Number()),
			},
		}, offset, limit)
		if err != nil {
			logger.Sugar().Errorw("CreateNotif", "error", err.Error())
			return
		}
		offset += limit
		if len(templateInfos) == 0 {
			logger.Sugar().Errorw("CreateNotif", "error", "template not exist")
			return
		}

		notifReq := []*notifmgrpb.NotifReq{}
		useTemplate := true
		date := time.Now().Format("2006-01-02")
		time1 := time.Now().Format("15:04:05")

		for key := range templateInfos {
			content := thirdpkg.ReplaceVariable(
				templateInfos[key].Content,
				userName,
				nil,
				amount,
				coinUnit,
				&date,
				&time1,
				address,
			)
			notifReq = append(notifReq, &notifmgrpb.NotifReq{
				AppID:       &appID,
				UserID:      &userID,
				LangID:      &templateInfos[key].LangID,
				EventType:   &eventType,
				UseTemplate: &useTemplate,
				Title:       &templateInfos[key].Title,
				Content:     &content,
				Channels:    []channelpb.NotifChannel{channelpb.NotifChannel_ChannelEmail},
			})
		}

		_, err = notifcli.CreateNotifs(ctx, notifReq)
		if err != nil {
			logger.Sugar().Errorw("CreateNotif", "error", err.Error())
			return
		}
	}
}

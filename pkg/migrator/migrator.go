//nolint
package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/NpoolPlatform/review-manager/pkg/db"
	"github.com/NpoolPlatform/review-manager/pkg/db/ent"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	reviewpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	ent1 "github.com/NpoolPlatform/review-service/pkg/db/ent"
	entreview "github.com/NpoolPlatform/review-service/pkg/db/ent/review"
	reviewconst "github.com/NpoolPlatform/review-service/pkg/message/const"

	"github.com/NpoolPlatform/go-service-framework/pkg/config"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	constant "github.com/NpoolPlatform/go-service-framework/pkg/mysql/const"

	_ "github.com/NpoolPlatform/review-service/pkg/db/ent/runtime"
)

const (
	keyUsername = "username"
	keyPassword = "password"
	keyDBName   = "database_name"
	maxOpen     = 10
	maxIdle     = 10
	MaxLife     = 3
)

func dsn(hostname string) (string, error) {
	username := config.GetStringValueWithNameSpace(constant.MysqlServiceName, keyUsername)
	password := config.GetStringValueWithNameSpace(constant.MysqlServiceName, keyPassword)
	dbname := config.GetStringValueWithNameSpace(hostname, keyDBName)

	svc, err := config.PeekService(constant.MysqlServiceName)
	if err != nil {
		logger.Sugar().Warnw("dsb", "error", err)
		return "", err
	}

	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true&interpolateParams=true",
		username, password,
		svc.Address,
		svc.Port,
		dbname,
	), nil
}

func open(hostname string) (conn *sql.DB, err error) {
	hdsn, err := dsn(hostname)
	if err != nil {
		return nil, err
	}

	logger.Sugar().Infow("open", "hdsn", hdsn)

	conn, err = sql.Open("mysql", hdsn)
	if err != nil {
		return nil, err
	}

	// https://github.com/go-sql-driver/mysql
	// See "Important settings" section.

	conn.SetConnMaxLifetime(time.Minute * MaxLife)
	conn.SetMaxOpenConns(maxOpen)
	conn.SetMaxIdleConns(maxIdle)

	return conn, nil
}

func migrateReview(ctx context.Context, conn *sql.DB) error {
	cli1 := ent1.NewClient(ent1.Driver(entsql.OpenDB(dialect.MySQL, conn)))
	rvs, err := cli1.
		Review.
		Query().
		Where(
			entreview.DeleteAt(0),
		).
		All(ctx)
	if err != nil {
		logger.Sugar().Errorw("migrateReview", "error", err)
		return err
	}

	return db.WithClient(ctx, func(_ctx context.Context, cli *ent.Client) error {
		_rvs, err := cli.
			Review.
			Query().
			Limit(1).
			All(_ctx)
		if err != nil {
			return err
		}
		if len(_rvs) > 0 {
			return nil
		}

		for _, rv := range rvs {
			objectType := rv.ObjectType
			domain := rv.Domain

			switch objectType {
			case "kyc":
				objectType = reviewpb.ReviewObjectType_ObjectKyc.String()
				domain = "appuser-gateway.npool.top"
			case "withdraw":
				objectType = reviewpb.ReviewObjectType_ObjectWithdrawal.String()
				domain = "ledger-gateway.npool.top"
			case reviewpb.ReviewObjectType_ObjectKyc.String():
				domain = "appuser-gateway.npool.top"
			case reviewpb.ReviewObjectType_ObjectWithdrawal.String():
				domain = "ledger-gateway.npool.top"
			default:
				continue
			}

			state := string(rv.State)
			switch state {
			case "approved":
				state = reviewpb.ReviewState_Approved.String()
			case "rejected":
				state = reviewpb.ReviewState_Rejected.String()
			case "wait":
				state = reviewpb.ReviewState_Wait.String()
			default:
				continue
			}

			trigger := rv.Trigger
			switch trigger {
			case "auto review":
				trigger = reviewpb.ReviewTriggerType_AutoReviewed.String()
			case "large amount":
				trigger = reviewpb.ReviewTriggerType_LargeAmount.String()
			case "insufficient":
				trigger = reviewpb.ReviewTriggerType_InsufficientFunds.String()
			}

			_, err := cli.
				Review.
				Create().
				SetID(rv.ID).
				SetAppID(rv.AppID).
				SetReviewerID(rv.ReviewerID).
				SetDomain(domain).
				SetObjectID(rv.ObjectID).
				SetObjectType(objectType).
				SetTrigger(trigger).
				SetState(state).
				SetMessage(rv.Message).
				Save(_ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func Migrate(ctx context.Context) error {
	if err := db.Init(); err != nil {
		logger.Sugar().Errorw("Migrate", "error", err)
		return err
	}

	conn, err := open(reviewconst.ServiceName)
	if err != nil {
		logger.Sugar().Errorw("Migrate", "error", err)
		return err
	}
	defer conn.Close()

	// Migrate review
	if err := migrateReview(ctx, conn); err != nil {
		logger.Sugar().Errorw("Migrate", "error", err)
		return err
	}

	return nil
}

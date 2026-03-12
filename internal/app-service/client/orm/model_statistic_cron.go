package orm

import (
	"context"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const (
	cornTaskModelStatisticSync = "CornTaskModelStatisticSync"
)

var (
	modelStatisticCronManager *ModelStatisticCronManager
)

type ModelStatisticCronManager struct {
	ctx  context.Context
	cron *cron.Cron
	db   *gorm.DB
}

func CronInit(ctx context.Context, db *gorm.DB) error {
	modelStatisticCronManager = &ModelStatisticCronManager{
		ctx:  ctx,
		cron: cron.New(),
		db:   db,
	}

	if err := syncModelStatistic(); err != nil {
		return fmt.Errorf("sync model statistic err: %v", err)
	}

	entryID, err := modelStatisticCronManager.cron.AddFunc("0 * * * *", cronSyncModelStatistic)
	if err != nil {
		log.Errorf("register cron task (%v) error: %v", cornTaskModelStatisticSync, err)
		return err
	}
	log.Infof("cron task (%v) registered with entry ID: %d", cornTaskModelStatisticSync, entryID)

	modelStatisticCronManager.cron.Start()

	return nil
}

func CronStop() {
	if modelStatisticCronManager != nil {
		modelStatisticCronManager.cron.Stop()
		log.Infof("cron tasks stopped")
	}
}

func cronSyncModelStatistic() {
	defer util.PrintPanicStack()
	if err := syncModelStatistic(); err != nil {
		log.Errorf("execute model statistic sync err: %v", err)
	}
}

func syncModelStatistic() error {
	ctx := modelStatisticCronManager.ctx
	db := modelStatisticCronManager.db

	now := time.Now().UnixMilli()
	startTs := now - 30*24*time.Hour.Milliseconds()
	dates := util.DateRange(startTs, now)

	for i := len(dates) - 1; i >= 0; i-- {
		date := dates[i]

		hasRecord, err := checkModelStatsRecordExists(ctx, db, date)
		if err != nil {
			return fmt.Errorf("check model stats record exists for date %v err: %v", date, err)
		}

		if err := updateModelStats(ctx, date, db); err != nil {
			log.Errorf("update model stats date %v err: %v", date, err)
		}

		if hasRecord {
			log.Infof("found existing record for date %v, stop backward sync", date)
			break
		}
	}
	return nil
}

func checkModelStatsRecordExists(ctx context.Context, db *gorm.DB, date string) (bool, error) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithDate(date),
	).Apply(db.WithContext(ctx)).Model(&model.ModelRecord{}).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check model stats record exists for date %v err: %v", date, err)
	}
	return count > 0, nil
}

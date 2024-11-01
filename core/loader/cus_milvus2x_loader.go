package loader

import (
	"context"
	"github.com/zilliztech/milvus-migration/core/common"
	"github.com/zilliztech/milvus-migration/core/config"
	"github.com/zilliztech/milvus-migration/core/dbclient"
	"github.com/zilliztech/milvus-migration/internal/log"
	"github.com/zilliztech/milvus-migration/internal/util/retry"
	"github.com/zilliztech/milvus-migration/storage/milvus2x"
	"go.uber.org/zap"
	"time"
)

// CustomMilvus2xLoader : can custom milvus collection info loader
type CustomMilvus2xLoader struct {
	CusMilvus2x *dbclient.CustomFieldMilvus2x
	concurLimit int
	workMode    string
	cfg         *config.MigrationConfig
	//jobId       string

	runtimeCollectionRows  map[string]int
	runtimeCollectionNames []string

	//custom fields collection info
	runtimeCusCollectionInfos []*common.CollectionInfo
}

func NewCusFieldMilvus2xLoader(cfg *config.MigrationConfig) (*CustomMilvus2xLoader, error) {
	client, err := dbclient.NewMilvus2xClient(cfg.TargetMilvus2xCfg)
	if err != nil {
		return nil, err
	}
	return &CustomMilvus2xLoader{
		CusMilvus2x: dbclient.NewCusFieldMilvus2xClient(client),
		cfg:         cfg,
		//jobId:       jobId,
		concurLimit: cfg.LoaderWorkLimit,
		workMode:    cfg.LoaderWorkCfg.WorkMode,
	}, nil
}

func (this *CustomMilvus2xLoader) Before(ctx context.Context) error {
	err := this.createTable(ctx)
	if err != nil {
		return err
	}
	return this.beforeStatistics(ctx)
}

func (this *CustomMilvus2xLoader) WriteByBulkInsert(ctx context.Context, fileName string, collection string) (int64, error) {

	log.LL(ctx).Info("[Loader] Begin to load json file to milvus",
		zap.String("collection", collection), zap.String("fileName", fileName))

	return this.CusMilvus2x.StartBulkLoad(ctx, collection, []string{fileName})
}

func (cus *CustomMilvus2xLoader) CheckMilvusState(ctx context.Context, taskId int64) error {
	//InBulkLoadProcess = errors.New("InBulkLoadProcess")
	//BulkLoadFailed    = errors.New("BulkLoadFailed")
	return cus.CusMilvus2x.CheckBulkLoadState(ctx, taskId)
}

func (cus *CustomMilvus2xLoader) After(ctx context.Context) error {
	return cus.compareResult(ctx)
}

func (cus *CustomMilvus2xLoader) createTable(ctx context.Context) error {

	log.LL(ctx).Info("[Loader] All collection Begin to create...")
	for _, collectionInfo := range cus.runtimeCusCollectionInfos {
		err := retry.Do(ctx, func() error {
			return cus.CusMilvus2x.CreateCollection(ctx, collectionInfo)
		}, retry.Attempts(5), retry.Sleep(2*time.Second))
		if err != nil {
			log.Error("fail to create collection after times", zap.String("collection", collectionInfo.Param.CollectionName), zap.Error(err))
			return err
		}
	}
	log.LL(ctx).Info("[Loader] Create All collection finish.")
	return nil
}

func (cus *CustomMilvus2xLoader) beforeStatistics(ctx context.Context) error {
	log.LL(ctx).Info("[Loader] collection load data before statistics...")
	colRowsMap, err := cus.CusMilvus2x.ShowCollectionRows(ctx, cus.runtimeCollectionNames, true)
	if err != nil {
		return err
	}
	cus.runtimeCollectionRows = colRowsMap
	return nil
}

func (this *CustomMilvus2xLoader) compareResult(ctx context.Context) error {
	rowsMap, err := this.CusMilvus2x.ShowCollectionRows(ctx, this.runtimeCollectionNames, false)

	beforeTotalCount := 0
	afterTotalCount := 0
	if err != nil {
		return err
	}
	for _, val := range this.runtimeCollectionNames {
		beforeCount := this.runtimeCollectionRows[val]
		afterCount := rowsMap[val]
		log.LL(ctx).Info("[Loader] Static: ", zap.String("collection", val),
			zap.Int("beforeCount", beforeCount),
			zap.Int("afterCount", afterCount),
			zap.Int("increase", afterCount-beforeCount))

		beforeTotalCount = beforeTotalCount + beforeCount
		afterTotalCount = afterTotalCount + afterCount
	}

	log.LL(ctx).Info("[Loader] Static Total", zap.Int("Total Collections", len(this.runtimeCollectionNames)),
		zap.Int("beforeTotalCount", beforeTotalCount),
		zap.Int("afterTotalCount", afterTotalCount),
		zap.Int("totalIncrease", afterTotalCount-beforeTotalCount))

	return nil
}

func (this *CustomMilvus2xLoader) BatchWrite(ctx context.Context, data *milvus2x.Milvus2xData) error {

	log.LL(ctx).Info("[Loader] Begin to batchWrite data to milvus", zap.String("collection",
		this.runtimeCollectionNames[0]), zap.String("partition", data.Partition))
	
	if this.cfg.TargetMilvus2xCfg.WriteMode == common.UPSERT {
		return this.CusMilvus2x.StartBatchUpsert(ctx, this.runtimeCollectionNames[0], data)
	} else {
		return this.CusMilvus2x.StartBatchInsert(ctx, this.runtimeCollectionNames[0], data)
	}
}

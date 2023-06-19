package reader

import (
	"github.com/zilliztech/milvus-migration/core/common"
	"github.com/zilliztech/milvus-migration/core/reader/source"
	"github.com/zilliztech/milvus-migration/core/transform/es/parser"
	"github.com/zilliztech/milvus-migration/internal/log"
	"go.uber.org/zap"
	"io"
	"strings"
	"time"
)

// ESReader :
type ESReader struct {
	//not read from file(local, remote), is from ES server, so don't use BaseReader params.
	//Cfg   *config.ESConfig
	BufSize  int
	ESSource *source.ESSource
}

func NewESReader(esSource *source.ESSource, bufSize int) *ESReader {
	esr := ESReader{
		BufSize:  bufSize,
		ESSource: esSource,
	}
	return &esr
}

func (esr *ESReader) BeforePublish() error {
	/*
		in others file/s3 Reader here need to get file/s3 io.reader
		but in es reader, just need es client, so here need to do nothing
			ps: create es client will happen the step of create ESSource.
	*/
	return nil
}

func (esr *ESReader) AfterPublish() error {
	/*
		in others file/s3 Reader here need to close file/s3 io.reader
		here in es reader we need close the es scroll
	*/
	return esr.ESSource.Close()
}

func (esr *ESReader) PublishTo(w io.Writer) error {
	defer log.Info("[ESReader] write ES success",
		zap.String("urls", strings.Join(esr.ESSource.Cfg.Urls, ",")),
		zap.String("cloudId", esr.ESSource.Cfg.CloudId),
		zap.String("version", esr.ESSource.Cfg.Version),
		//zap.String("security", esr.ESSource.Cfg.Security),
	)
	return esr.writeAll(w)
}

func (esr *ESReader) writeAll(w io.Writer) error {
	log.Info("[ESReader] begin to write json data...")
	start := time.Now()
	//1. write first from es source
	data, err := esr.ESSource.ReadFirst()
	if err != nil {
		log.Error("[ESReader] write json data", zap.Error(err))
		return err
	}
	if data.IsEmpty {
		log.Warn("[ESReader] end to write, json data is empty")
		return nil
	}
	b := esparser.ToMilvus2Format(data.Hits, true)
	w.Write(b)
	var batch = 0
	var printSize = 100
	//2. foreach write next data from es source
	for !data.IsEmpty {
		data, err = esr.ESSource.ReadNext()
		if err != nil {
			log.Error("[ESReader] foreach write json data", zap.Error(err))
			return err
		}
		if data.IsEmpty {
			break
		}

		var start0 time.Time
		if common.DEBUG {
			start0 = time.Now()
		}
		b = esparser.ToMilvus2Format(data.Hits, false)
		w.Write(b)

		if common.DEBUG {
			log.Info("[ESReader] 3 Es data parser to Writer ======>", zap.Float64("Cost", time.Since(start0).Seconds()))
		}

		if (batch % printSize) == 0 {
			log.Info("[ESReader] 4 writing batch es data =======> ", zap.Int("Batch", batch),
				zap.Float64("Cost", time.Since(start).Seconds()))
		}
		batch++
	}
	w.Write(esparser.EndCharacter())
	log.Info("[ESReader] success end to write json data=======>", zap.Float64("Cost", time.Since(start).Seconds()))
	return nil
}

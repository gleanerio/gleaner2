package pkg

import (
	"github.com/gleanerio/gleaner2/internal/common"
	"github.com/gleanerio/gleaner2/pkg/storage"
	"github.com/gleanerio/gleaner2/internal/services/bulk"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/minio/minio-go/v7"
)

func Bulk(v1 *viper.Viper, mc *minio.Client) error {
	//err := bulk.ObjectAssembly(v1, mc)
	err := bulk.BulkAssembly(v1, mc)

	if err != nil {
		log.Error(err)
	}
	return err
}

// Wrapper that builds its own minio client from the config.
// TODO: develop a common config for the services (s3, graph, etc.)
func NabuBulk(v1 *viper.Viper) error {
	common.InitLogging()
	mc, err := storage.MinioConnection(v1)
	if err != nil {
		log.Fatalf("cannot connect to minio: %s", err)
	}
	return Bulk(v1, mc)
}

package pkg

import (
	"github.com/gleanerio/nabu/internal/common"
	"github.com/gleanerio/nabu/internal/services/bulk"
	log "github.com/sirupsen/logrus"

	"github.com/gleanerio/nabu/pkg/storage"
	"github.com/spf13/viper"

	"github.com/minio/minio-go/v7"
)

func GraphDB(v1 *viper.Viper, mc *minio.Client) error {
	//err := bulk.ObjectAssembly(v1, mc)
	err := bulk.BulkAssembly(v1, mc)

	if err != nil {
		log.Error(err)
	}
	return err
}

// Wrapper that builds its own minio client from the config.
// TODO: develop a common config for the services (s3, graph, etc.)
func NabuGraphDB(v1 *viper.Viper) error {
	common.InitLogging()
	mc, err := storage.MinioConnection(v1)
	if err != nil {
		log.Fatalf("cannot connect to minio: %s", err)
	}
	return GraphDB(v1, mc)
}

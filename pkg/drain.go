package pkg

import (
	"github.com/gleanerio/nabu/internal/common"
	"github.com/gleanerio/nabu/pkg/storage"
	"github.com/gleanerio/nabu/internal/services/drain"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Drain(v1 *viper.Viper, mc *minio.Client) error {
	log.Info("Nabu started with mode: drain")
	err := drain.Objects(v1, mc)

	if err != nil {
		log.Error(err)
	}
	return err
}

// Wrapper that builds its own minio client from the config.
// TODO: develop a common config for the services (s3, graph, etc.)
func NabuDrain(v1 *viper.Viper) error {
	common.InitLogging()
	mc, err := storage.MinioConnection(v1)
	if err != nil {
		log.Fatalf("cannot connect to minio: %s", err)
	}
	return Drain(v1, mc)
}

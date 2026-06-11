package pkg

import (
	"github.com/gleanerio/gleaner2/internal/common"
	"github.com/gleanerio/gleaner2/pkg/storage"
	"github.com/gleanerio/gleaner2/internal/prune"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Prune(v1 *viper.Viper, mc *minio.Client) error {
	log.Info("Prune graphs in triplestore not in objectVal store")
	err := prune.Snip(v1, mc)
	if err != nil {
		log.Error(err)
	}
	return err
}

// Wrapper that builds its own minio client from the config.
// TODO: develop a common config for the services (s3, graph, etc.)
func NabuPrune(v1 *viper.Viper) error {
	common.InitLogging()
	mc, err := storage.MinioConnection(v1)
	if err != nil {
		log.Fatalf("cannot connect to minio: %s", err)
	}
	return Prune(v1, mc)
}

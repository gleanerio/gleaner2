package pkg

import (
	"github.com/gleanerio/gleaner2/internal/common"
	"github.com/gleanerio/gleaner2/internal/sparql"
	"github.com/gleanerio/gleaner2/pkg/storage"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Clear(v1 *viper.Viper, mc *minio.Client) error {
	log.Info("Nabu started with mode: clear")

	d := v1.GetBool("flags.dangerous")

	if d {
		log.Println("dangerous mode is enabled")
		_, err := sparql.Clear(v1)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		log.Println("dangerous mode must be set to true to run this")
		return nil
	}

	return nil
}

// NabuClear is a wrapper that builds its own minio client from the config.
// TODO: develop a common config for the services (s3, graph, etc.)
func NabuClear(v1 *viper.Viper) error {
	common.InitLogging()
	mc, err := storage.MinioConnection(v1)
	if err != nil {
		log.Fatalf("cannot connect to minio: %s", err)
	}
	return Clear(v1, mc)
}

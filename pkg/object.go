package pkg

import (
	"fmt"
	"github.com/gleanerio/gleaner2/internal/common"
	"github.com/gleanerio/gleaner2/pkg/storage"
	"github.com/gleanerio/gleaner2/internal/services/bulk"
//	"github.com/gleanerio/gleaner2/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/minio/minio-go/v7"
)

func Object(v1 *viper.Viper, mc *minio.Client, bucket string, object string) error {
	fmt.Println("Load graph object to triplestore")
//	spql, _ := config.GetSparqlConfig(v1)
//	if bucket == "" {
//		bucket, _ = config.GetBucketName(v1)
//	}
//	s, err := objects.PipeLoad(v1, mc, bucket, object, spql.Endpoint)
//	if err != nil {
//		log.Error(err)
//	}

	s, err := bulk.BulkLoad(v1, mc, bucket, object)
	if err != nil {
		log.Println(err)
	}

	log.Trace(string(s))
	return err
}

// Wrapper that builds its own minio client from the config.
// TODO: develop a common config for the services (s3, graph, etc.)
func NabuObject(v1 *viper.Viper, bucket string, object string) error {
	common.InitLogging()
	mc, err := storage.MinioConnection(v1)
	if err != nil {
		log.Fatalf("cannot connect to minio: %s", err)
	}
	return Object(v1, mc, bucket, object)
}

package common

// Bridge functions that delegate to pkg/graph and pkg/storage for backward
// compatibility with Gleaner code that uses common.JLDProc, common.JLD2nq,
// common.MinioConnection, and common.PipeCopyNG.

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/minio/minio-go/v7"
	"github.com/piprate/json-gold/ld"
	"github.com/spf13/viper"

	"github.com/gleanerio/nabu/pkg/graph"
	"github.com/gleanerio/nabu/pkg/storage"
)

// JLDProc delegates to pkg/graph.JLDProc for JSON-LD processor creation
func JLDProc(v1 *viper.Viper) (*ld.JsonLdProcessor, *ld.JsonLdOptions) {
	return graph.JLDProc(v1)
}

// JLD2nq converts JSON-LD documents to NQuads using a pre-built processor.
// This signature takes pre-built proc/options for batch processing efficiency.
func JLD2nq(jsonld string, proc *ld.JsonLdProcessor, options *ld.JsonLdOptions) (string, error) {
	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Error(err)
		return "", err
	}

	nq, err := proc.ToRDF(myInterface, options)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return nq.(string), err
}

// MinioConnection provides Gleaner-compatible MinIO connection (panics on error).
// Gleaner's original returned *minio.Client without error.
func MinioConnection(v1 *viper.Viper) *minio.Client {
	mc, err := storage.MinioConnection(v1)
	if err != nil {
		log.Fatal(err)
	}
	return mc
}

// PipeCopyNG concatenates objects from a prefix into a single object.
// This is the Gleaner-compatible version that delegates to a simple pipe copy.
func PipeCopyNG(name, bucket, prefix string, mc *minio.Client) error {
	log.Debug("Start pipe reader / writer sequence")

	pr, pw := io.Pipe()
	lwg := sync.WaitGroup{}
	lwg.Add(2)

	isRecursive := true

	go func() {
		defer lwg.Done()
		defer pw.Close()
		objectCh := mc.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive})

		for object := range objectCh {
			fo, err := mc.GetObject(context.Background(), bucket, object.Key, minio.GetObjectOptions{})
			if err != nil {
				fmt.Println(err)
			}

			var b bytes.Buffer
			bw := bufio.NewWriter(&b)

			_, err = io.Copy(bw, fo)
			if err != nil {
				log.Error(err)
			}

			pw.Write(b.Bytes())
		}
	}()

	go func() {
		defer lwg.Done()
		_, err := mc.PutObject(context.Background(), bucket, name, pr, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Error(err)
		}
	}()

	lwg.Wait()
	pw.Close()
	pr.Close()

	return nil
}

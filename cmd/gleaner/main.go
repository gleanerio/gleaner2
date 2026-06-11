package main

import (
	"github.com/gleanerio/gleaner2/internal/common"
	"github.com/gleanerio/gleaner2/pkg/cli"
)

func init() {
	common.InitLogging()
}

func main() {
	cli.Execute()
}

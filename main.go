package main

import (
	aivenUtils "github.com/jslauthor/infra-aiven/pkg/aiven"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		err := aivenUtils.DeployAiven(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}

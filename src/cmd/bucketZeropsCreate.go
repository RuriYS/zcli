package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/zerops-io/zcli/src/cliAction/bucket/zerops"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
)

func bucketZeropsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create projectNameOrId serviceName bucketName [flags]",
		Short:        i18n.CmdBucketCreate,
		Args:         ExactNArgs(3),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			xAmzAcl, err := getXAmzAcl(cmd)
			if err != nil {
				return err
			}

			storage, err := createCliStorage()
			if err != nil {
				return err
			}

			token, err := getToken(storage)
			if err != nil {
				return err
			}

			region, err := createRegionRetriever(ctx)
			if err != nil {
				return err
			}

			reg, err := region.RetrieveFromFile()
			if err != nil {
				return err
			}

			apiClientFactory := zBusinessZeropsApiProtocol.New(zBusinessZeropsApiProtocol.Config{
				CaCertificateUrl: reg.CaCertificateUrl,
			})
			apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(
				ctx,
				reg.GrpcApiAddress,
				token,
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			client := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			b := bucketZerops.New(bucketZerops.Config{}, client, apiGrpcClient, sdkConfig.Config{Token: token, RegionUrl: reg.RestApiAddress})
			return b.Create(ctx, bucketZerops.RunConfig{
				ProjectNameOrId:  args[0],
				ServiceStackName: args[1],
				BucketName:       args[2],
				XAmzAcl:          xAmzAcl,
			})
		},
	}
	params.RegisterString(cmd, "x-amz-acl", "", i18n.BucketGenericXAmzAcl)

	cmd.Flags().BoolP("help", "h", false, helpText(i18n.BucketCreateHelp))
	return cmd
}
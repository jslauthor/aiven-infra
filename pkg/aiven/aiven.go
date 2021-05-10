package aivenUtils

import (
	"fmt"

	"github.com/linkedin/goavro/v2"
	"github.com/pulumi/pulumi-aiven/sdk/v4/go/aiven"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func DeployAiven(ctx *pulumi.Context) error {

	cfg := config.New(ctx, "")

	/* Retrieve Project in Aiven  -- can also create one from scratch */

	project, err := aiven.LookupProject(ctx, &aiven.LookupProjectArgs{
		Project: "core",
	}, nil)
	if err != nil {
		return err
	}

	/* Create Kafka cluster */

	kafka, err := aiven.NewKafka(ctx, "aiven-kafka-dev", &aiven.KafkaArgs{
		Project:               pulumi.Sprintf("%s", project.Id),
		CloudName:             pulumi.String("aws-us-west-2"),
		Plan:                  pulumi.String("startup-2"),
		ServiceName:           pulumi.String("kafka-dev"),
		MaintenanceWindowDow:  pulumi.String("monday"),
		MaintenanceWindowTime: pulumi.String("10:00:00"),
		KafkaUserConfig: &aiven.KafkaKafkaUserConfigArgs{
			KafkaRest:      pulumi.String("true"),
			KafkaConnect:   pulumi.String("false"),
			SchemaRegistry: pulumi.String("true"),
			KafkaVersion:   pulumi.String("2.8"),
			Kafka: &aiven.KafkaKafkaUserConfigKafkaArgs{
				GroupMaxSessionTimeoutMs: pulumi.String("70000"),
				LogRetentionBytes:        pulumi.String("1000000000"),
			},
			PublicAccess: &aiven.KafkaKafkaUserConfigPublicAccessArgs{
				KafkaRest:    pulumi.String("true"),
				KafkaConnect: pulumi.String("true"),
			},
			KafkaAuthenticationMethods: aiven.KafkaKafkaUserConfigKafkaAuthenticationMethodsArgs{
				Sasl:        pulumi.String("true"),
				Certificate: pulumi.String("true"),
			},
		},
	})
	if err != nil {
		return err
	}

	/* Get and Create Schemas */

	// We would place all of the codecs in a common go libarary shared across all go projects
	chatCodec, err := goavro.NewCodec(`
	{
		"name": "ChatMessage",
		"type": "record",
		"fields": [
			{
				"type": "string",
				"name": "socket_id"
			},
			{
				"type": "string",
				"name": "channel_id"
			},
			{
				"type": "string",
				"name": "message_id"
			},
			{
				"type": "string",
				"name": "user_id"
			},
			{
				"type": "string",
				"name": "group_id"
			},
			{
				"type": "string",
				"name": "message"
			},
			{
				"type": "string",
				"name": "timestamp"
			}
	
		]
	}`)
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("%s", err), nil)
	}

	_, err = aiven.NewKafkaSchema(ctx, "kafka-chat-schema", &aiven.KafkaSchemaArgs{
		Project:            pulumi.Sprintf("%s", project.Id),
		ServiceName:        kafka.ServiceName,
		SubjectName:        pulumi.String("kafka-chat-schema"),
		CompatibilityLevel: pulumi.String("FORWARD"),
		Schema:             pulumi.String(chatCodec.Schema()),
	})
	if err != nil {
		return err
	}

	/* Create Topics */

	_, err = aiven.NewKafkaTopic(ctx, "kafka-chat-topic", &aiven.KafkaTopicArgs{
		Project:     pulumi.Sprintf("%s", project.Id),
		ServiceName: kafka.ServiceName,
		Partitions:  pulumi.Int(cfg.RequireInt("numberOfPartitions")),
		Replication: pulumi.Int(2),
		TopicName:   pulumi.String("aws.user.chat.cmd.v1"),
	})
	if err != nil {
		return err
	}

	/* Outputs */

	ctx.Export("KafkaUrl", pulumi.Sprintf("%s", kafka.ServiceUri))
	ctx.Export("KafkaRestUri", kafka.Kafka.RestUri().ApplyT(func(url *string) string {
		return *url
	}))

	return nil

}

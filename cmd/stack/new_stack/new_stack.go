package newstack

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type PixifyStackProps struct {
	awscdk.StackProps
}

func NewPixifyStack(scope constructs.Construct, id string, props *PixifyStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// #region //*====== Create Raw S3 Bucket ========

	rawImageBucket := awss3.NewBucket(stack, jsii.String("PixifyRawImagesBucket"), &awss3.BucketProps{
		Versioned:         jsii.Bool(true),
		BucketName:        jsii.String("pixify-raw-images-bucket"),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		EnforceSSL:        jsii.Bool(true),
		AutoDeleteObjects: jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
	})

	awscdk.NewCfnOutput(stack, jsii.String("RawImagesBucket"), &awscdk.CfnOutputProps{
		Value:       rawImageBucket.BucketName(),
		Description: jsii.String("S3 bucket to store raw images"),
	})

	// #endregion //*====== Create Raw S3 Bucket ========

	// #region //*====== Create Transform S3 Bucket ========

	transformedImageBucket := awss3.NewBucket(stack, jsii.String("PixifyTransformedImagesBucket"), &awss3.BucketProps{
		BucketName:        jsii.String("pixify-transformed-images-bucket"),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		EnforceSSL:        jsii.Bool(true),
		AutoDeleteObjects: jsii.Bool(true),
	})

	awscdk.NewCfnOutput(stack, jsii.String("TransformedImagesBucket"), &awscdk.CfnOutputProps{
		Value:       transformedImageBucket.BucketName(),
		Description: jsii.String("S3 bucket to store transformed images"),
	})

	// #endregion //*====== Create Transform S3 Bucket ========

	// #region //*====== Create Lambda function ========

	transformerFunc := awslambda.NewFunction(stack, jsii.String("TransformerLambdaFunc"), &awslambda.FunctionProps{
		FunctionName: jsii.String("pixify-transformer"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap.handler"),
		Environment: &map[string]*string{
			"originalImageBucketName":  rawImageBucket.BucketName(),
			"transformImageBucketName": transformedImageBucket.BucketName(),
		},
		LogRetention: awslogs.RetentionDays_ONE_DAY,
		Architecture: awslambda.Architecture_ARM_64(),

		Code: awslambda.AssetCode_FromAsset(jsii.String("./../lambda/go_lambda.zip"), &awss3assets.AssetOptions{}),
	})

	lambdaUrl := transformerFunc.AddFunctionUrl(&awslambda.FunctionUrlOptions{})

	rawImageBucket.GrantRead(transformerFunc, "*")
	transformedImageBucket.GrantWrite(transformerFunc, "*", nil)
	s3RawImagePolicy := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: &[]*string{
			jsii.String("s3:GetObject"),
			jsii.String("s3:ListBucket"),
		},
		Resources: &[]*string{
			jsii.String("arn:aws:s3:::pixify-raw-images-bucket/*"),
		},
	})

	s3TransformedImagePolicy := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: &[]*string{
			jsii.String("s3:PutObject"),
		},
		Resources: &[]*string{
			jsii.String(fmt.Sprintf("arn:aws:s3:::%v/*", *transformedImageBucket.BucketName())),
		},
	})

	transformerFunc.Role().AttachInlinePolicy(awsiam.NewPolicy(stack, jsii.String("ReadOriginalANDPutTransformedImages"), &awsiam.PolicyProps{
		Statements: &[]awsiam.PolicyStatement{
			s3RawImagePolicy,
			s3TransformedImagePolicy,
		},
	}))

	awscdk.NewCfnOutput(stack, jsii.String("LambdaFunctionURL"), &awscdk.CfnOutputProps{
		Value:       lambdaUrl.Url(),
		Description: jsii.String("Lambda func to transform images"),
	})

	// #endregion //*====== Create Lambda function ========

	// #region //*====== Create Cloudfront distribution ========

	cloudFrontOriginGroup := awscloudfrontorigins.NewOriginGroup(&awscloudfrontorigins.OriginGroupProps{
		PrimaryOrigin:  awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(transformedImageBucket, nil),
		FallbackOrigin: awscloudfrontorigins.FunctionUrlOrigin_WithOriginAccessControl(lambdaUrl, nil),
		FallbackStatusCodes: &[]*float64{
			jsii.Number(403),
			jsii.Number(500),
			jsii.Number(503),
			jsii.Number(504),
		},
	})

	urlRewriteFunc := awscloudfront.NewFunction(stack, jsii.String("UrlRewriteFunction"), &awscloudfront.FunctionProps{
		Code: awscloudfront.FunctionCode_FromFile(&awscloudfront.FileCodeOptions{
			FilePath: jsii.String("./url_rewrite/urlRewrite.js"),
		}),
		FunctionName: jsii.String("urlRewriteFunction"),
		Runtime:      awscloudfront.FunctionRuntime_JS_2_0(),
	})
	cloudFrontDist := awscloudfront.NewDistribution(stack, jsii.String("CloudfrontDistribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: cloudFrontOriginGroup,
			FunctionAssociations: &[]*awscloudfront.FunctionAssociation{
				{
					EventType: awscloudfront.FunctionEventType_VIEWER_REQUEST,
					Function:  urlRewriteFunc,
				},
			},
			ResponseHeadersPolicy: awscloudfront.NewResponseHeadersPolicy(stack, jsii.String("CloudfrontResponseHeaderPolicy"), &awscloudfront.ResponseHeadersPolicyProps{

				CorsBehavior: &awscloudfront.ResponseHeadersCorsBehavior{
					AccessControlAllowCredentials: jsii.Bool(false),
					AccessControlAllowHeaders: &[]*string{
						jsii.String("*"),
					},
					AccessControlAllowMethods: &[]*string{
						jsii.String("GET"),
					},
					AccessControlAllowOrigins: &[]*string{
						jsii.String("*"),
					},
					AccessControlMaxAge: awscdk.Duration_Seconds(jsii.Number(600)),
					OriginOverride:      jsii.Bool(false),
				},
			}),

			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			CachePolicy: awscloudfront.NewCachePolicy(stack, jsii.String("CloudfrontCachePolicy"), &awscloudfront.CachePolicyProps{
				DefaultTtl:          awscdk.Duration_Days(jsii.Number(24)),
				MaxTtl:              awscdk.Duration_Days(jsii.Number(365)),
				MinTtl:              awscdk.Duration_Seconds(jsii.Number(0)),
				QueryStringBehavior: awscloudfront.CacheQueryStringBehavior_All(),
			}),
		},
		Comment: jsii.String("cloudfrontDistributionFunction"),
	})

	awscloudfront.NewCfnOriginAccessControl(stack, jsii.String("OACLambda"), &awscloudfront.CfnOriginAccessControlProps{
		OriginAccessControlConfig: &awscloudfront.CfnOriginAccessControl_OriginAccessControlConfigProperty{
			Name:                          jsii.String("lambda-oac"),
			Description:                   jsii.String("Origin Access Control"),
			OriginAccessControlOriginType: jsii.String("lambda"),
			SigningBehavior:               jsii.String("always"),
			SigningProtocol:               jsii.String("sigv4"),
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("cloudFrontURL"), &awscdk.CfnOutputProps{
		Value: cloudFrontDist.DistributionDomainName(),
	})

	return stack
}

import * as path from "path";
import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import * as acm from "aws-cdk-lib/aws-certificatemanager";
import * as apigwv2integration from "@aws-cdk/aws-apigatewayv2-integrations-alpha";
import * as apigwv2 from "@aws-cdk/aws-apigatewayv2-alpha";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as logs from "aws-cdk-lib/aws-logs";
import * as route53 from "aws-cdk-lib/aws-route53";
import * as targets from "aws-cdk-lib/aws-route53-targets";

export class InfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const bsdlpDevZone = route53.HostedZone.fromHostedZoneId(
      this,
      "bsdlpDevHostedZone",
      "Z05432021TEFY2IZX7IFD"
    );

    const serverLambda = new lambda.Function(this, "serverLambda", {
      code: lambda.Code.fromAsset(path.join(__dirname, "../../build/")),
      runtime: lambda.Runtime.GO_1_X,
      handler: "server",
      logRetention: logs.RetentionDays.THREE_DAYS,
    });

    const lambdaIntegration = new apigwv2integration.HttpLambdaIntegration(
      "server",
      serverLambda,
      {
        payloadFormatVersion: apigwv2.PayloadFormatVersion.VERSION_1_0,
      }
    );

    const cert = new acm.Certificate(this, "certificate", {
      domainName: "dataurl.bsdlp.dev",
      validation: acm.CertificateValidation.fromDns(bsdlpDevZone),
    });

    const dn = new apigwv2.DomainName(this, "DN", {
      domainName: "dataurl.bsdlp.dev",
      certificate: acm.Certificate.fromCertificateArn(
        this,
        "cert",
        cert.certificateArn
      ),
    });

    const httpApi = new apigwv2.HttpApi(this, "apigateway", {
      defaultDomainMapping: {
        domainName: dn,
      },
    });
    httpApi.addRoutes({
      path: "/",
      methods: [apigwv2.HttpMethod.POST, apigwv2.HttpMethod.GET],
      integration: lambdaIntegration,
    });

    new route53.ARecord(this, "gatewayAliasRecord", {
      zone: bsdlpDevZone,
      recordName: "dataurl.bsdlp.dev.",
      target: route53.RecordTarget.fromAlias(
        new targets.ApiGatewayv2DomainProperties(
          dn.regionalDomainName,
          dn.regionalHostedZoneId
        )
      ),
    });
  }
}

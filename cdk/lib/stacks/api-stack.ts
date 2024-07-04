import { join } from 'path';
import { Duration, Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Code, Function, Runtime } from 'aws-cdk-lib/aws-lambda';
import { HttpLambdaIntegration } from 'aws-cdk-lib/aws-apigatewayv2-integrations';
import { CorsHttpMethod, HttpApi, HttpMethod } from 'aws-cdk-lib/aws-apigatewayv2';
import { TableV2 } from 'aws-cdk-lib/aws-dynamodb';

interface ApiStackProps extends StackProps {
  ddbTable: TableV2;
}

export class ApiStack extends Stack {

  constructor(scope: Construct, id: string, props: ApiStackProps) {
    super(scope, id, props);

    const demoLambda = new Function(this, 'DemoLambda', {
      runtime: Runtime.PROVIDED_AL2023,
      code: Code.fromAsset(join(__dirname, '../../../'), {
        bundling: {
          image: Runtime.PROVIDED_AL2023.bundlingImage,
          user: 'root',
          command: [
            "/bin/sh",
            "-c",
            "GOOS=linux go build -tags lambda.norpc -o /asset-output/bootstrap /asset-input/src/main.go"
          ],
          // NOTE: Can mount local  repo to avoid re-downloading all the dependencies. Maven ex:
          // volumes: [
          //   {
          //     hostPath: join(homedir(), '.m2'),
          //     containerPath: '/root/.m2/'
          //   }
          // ],
          // outputType: BundlingOutput.ARCHIVED
        }
      }),
      handler: 'bootstrap',
      environment: {
        DDB_TABLE_NAME: props.ddbTable.tableName,
      },
      timeout: Duration.seconds(7),
    });

    props.ddbTable.grantReadWriteData(demoLambda);

    const demoLambdaIntegration =
      new HttpLambdaIntegration('DemoLambdaIntegration', demoLambda);

    const httpApi = new HttpApi(this, 'HttpApi', {
      createDefaultStage: false,
      corsPreflight: {
        allowHeaders: ['Authorization'],
        allowMethods: [CorsHttpMethod.ANY],
        allowOrigins: ['*'],
        maxAge: Duration.days(10),
      },
    });

    httpApi.addStage('DefaultStage', {
      stageName: '$default',
      autoDeploy: true,
      throttle: {
        burstLimit: 2,
        rateLimit: 1,
      }
    });

    httpApi.addRoutes({
      path: '/users',
      methods: [HttpMethod.GET, HttpMethod.POST],
      integration: demoLambdaIntegration,
    });

    httpApi.addRoutes({
      path: '/users/{id}',
      methods: [HttpMethod.GET, HttpMethod.PUT, HttpMethod.DELETE],
      integration: demoLambdaIntegration,
    });
  }
}

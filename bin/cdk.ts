#!/usr/bin/env node

import { App } from 'aws-cdk-lib';

import { DbStack } from '../cdk/lib/stacks/db-stack';

import { ApiStack } from '../cdk/lib/stacks/api-stack';

const app = new App();

const dbStack = new DbStack(app, 'DbStack', {});

const apiStack = new ApiStack(app, 'ApiStack', {
  ddbTable: dbStack.ddbTable,
});

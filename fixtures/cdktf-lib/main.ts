import { provider, dataNullDataSource } from "@cdktf/provider-null";
import { Construct } from "constructs";
import {
  TerraformStack,
  TerraformElement,
  HttpBackend,
  HttpBackendConfig,
} from "cdktf";

export class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string, backend: HttpBackendConfig) {
    super(scope, id);
    new HttpBackend(this, backend);
    new provider.NullProvider(this, "null", {});
  }
}

export class MyResource extends TerraformElement {
  constructor(scope: Construct, id: string) {
    super(scope, id);
    new dataNullDataSource.DataNullDataSource(this, "Resource", {
      inputs: {
        example: "example",
      },
    });
  }
}

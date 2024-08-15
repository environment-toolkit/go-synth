import { Construct } from "constructs";
import { TerraformStack, TerraformElement, HttpBackendConfig } from "cdktf";
export declare class MyStack extends TerraformStack {
    constructor(scope: Construct, id: string, backend: HttpBackendConfig);
}
export declare class MyResource extends TerraformElement {
    constructor(scope: Construct, id: string);
}

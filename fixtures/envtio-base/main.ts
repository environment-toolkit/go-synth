import { App, HttpBackend } from "cdktf";
import { aws } from "@envtio/base";

const outdir = "cdktf.out";
const app = new App({
  outdir,
});
const stack = new aws.AwsSpec(app, "sample-stack", {
  providerConfig: {
    region: "us-west-2",
  },
});

new aws.network.SimpleIPv4(stack, "network", {
  gridUUID: "12345678-1234",
  environmentName: "test",
  config: {
    internalDomain: "example.com",
    ipv4CidrBlock: "10.0.0.0/16",
  },
});

new HttpBackend(stack, {
  address: "http://localhost:1234",
});

app.synth();

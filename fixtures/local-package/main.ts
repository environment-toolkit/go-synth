// import * as path from "path";
import { App } from "cdktf";
import { MyStack, MyResource } from "cdktf-lib";

const outdir = "cdktf.out";
const app = new App({
  outdir,
});
const stack = new MyStack(app, "my-stack", {
  address: "http://localhost:1234",
});
new MyResource(stack, "my-resource");

app.synth();

// const resultPath = path.join(
//   outdir,
//   app.manifest.forStack(stack).synthesizedStackPath
// );

// // return the synthesized stack
// const result = await Bun.file(resultPath).text();

// // if running in a child process, send the result back to the parent
// if (process.send) process.send(result);
// else console.log(result);

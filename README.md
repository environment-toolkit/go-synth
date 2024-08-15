# go-synth

A Go library to dynamically synth Constructs.

Given a CDKTF Entry script, return the directory with synthesized output

## BunExecutor

> [!WARNING]
> Bun binary is expected on `PATH`

<!-- todo: check if bun is on PATH -->

Example usage (see [cmd/main.go](./cmd/main.go))

```golang
logger := zap.NewProduction()
app := synth.NewApp(executors.NewBunExecutor, logger)
app.Configure(ctx, config.App{
    Dependencies: map[string]string{
        "cdktf": "^0.20.7",
    },
    DevDependencies: map[string]string{
        "@types/bun": "^1.1.3",
        "typescript": "5.4.5",
    },
    ScopedPackages: []config.ScopedPackageOptions{},
})
// prepare afero fs to receive the result
dest := afero.NewOsFs()

mainTs := `import { App } from "cdktf";
import { MyStack, MyResource } from "my-cdktf-pkg";

const outdir = "cdktf.out";
const app = new App({
  outdir,
});
const stack = new MyStack(app, "my-stack", {
  address: "http://localhost:1234",
});
new MyResource(stack, "my-resource");

app.synth();`

// Execute the main.ts script
app.Eval(ctx, mainTs, "cdktf.out/stacks/my-stack", ".", dest)
```

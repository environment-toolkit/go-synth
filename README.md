# go-synth

A Go library to dynamically synth Constructs.

Given a CDKTF App "entry" script, return synthesized output.

## BunExecutor

> [!WARNING]
> Bun binary is expected on `PATH`

<!-- TODO: check if bun is on PATH? -->

Example usage (see [cmd/main.go](./cmd/main.go))

```golang
logger := zap.NewProduction()
app := synth.NewApp(executors.NewBunExecutor, logger)
app.Configure(ctx, models.AppConfig{
    Dependencies: map[string]string{
      "my-cdktf-pkg": "0.0.1",
    },
    ScopedPackages: []models.ScopedPackageOptions{},
})
// prepare afero fs to receive the result
dstFs := afero.NewOsFs()

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

// Execute the main.ts script and copy the synthesized stack out
if err = app.Eval(ctx, dstFs, string(mainTs), *srcDir, *outDir); err != nil {
  logger.Fatal("Failed to execute main.ts script", zap.Error(err))
}
```

## NodeExecutor

> [!WARNING]
> Requires valid `NodeJS` and package manager (default: `pnpm`) on `$PATH`

Example usage (see [executors/node_executor_test.go](./executors/node_executor_test.go))

```golang
logger := zap.NewProduction()
app := synth.NewApp(executors.NewNodeExecutor, logger)
app.Configure(ctx, models.AppConfig{
    Dependencies: map[string]string{
      "my-cdktf-pkg": "0.0.1",
    },
    DevDependencies: map[string]string{
      "@swc/core":  "^1.7.6", // swc is included by default
    },
    ExecutorOptions: map[string]string{
      // entrypoint setup (install) and eval (run)
      "entrypoint": "pnpm",
      // script ran by pnpm to synth main.ts
      "synthScript": "ts-node --swc -P ./tsconfig.json main.ts",
    },
})
// prepare afero fs to receive the result
destFs := afero.NewOsFs()

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

// Execute the main.ts script and copy the synthesized stack out
if err = app.Eval(ctx, dstFs, string(mainTs), *srcDir, *outDir); err != nil {
  logger.Fatal("Failed to execute main.ts script", zap.Error(err))
}
```

## FAQ

### JSII supports Golang, what is this?

While it is true Constructs may be written in Golang or TS Constructs can be cross-compiled to Golang using JSII, dynamically loading these libraries and executing them is not possible.

`go-synth` addresses this by allowing you to dynamically generate the CDKTF app and synthesize it.

### Why doesn't it work on my machine?

This library has currently only been manually test on Linux (WSL) with Bun 1.1.23, NodeJS v20 and pnpm v9.

Feel free to open a ticket until CI/CD has been configured.

## Todo

- [x] Add Bun Executor
- [x] Add Node+Pnpm Executor
- [x] Add LICENSE
- [ ] Add CI/CD and Release process
- [ ] Add golang executor (use JSII cross compiled CDKTF Golang constructs with `go run`)
- [ ] Add go-typescript executor ([goja/#519(comment)](https://github.com/dop251/goja/issues/519#issuecomment-1592935649) / [go-typescript/pull/13](https://github.com/clarkmcc/go-typescript/pull/13))

Add benchmarking across executors.

# Fixtures

Sample configurations to test `go-synth` against.

## local-package

Sample `main.ts` referencing a local library (used in executor tests)

## TerraConstructs/Base

Sample `main.ts` creating an AWS Network with [@terraconstructs/base](https://www.npmjs.com/package/@terraconstructs/base)

```console
synth -file main.ts -deps "@terraconstructs/base:0.1.0" -src "cdktf.out/stacks/sample-stack"
```

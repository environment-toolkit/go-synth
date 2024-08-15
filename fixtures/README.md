# Fixtures

Sample configurations to test `go-synth` against.

## local-package

Sample `main.ts` referencing a local library (used in executor tests)

## Envtio/Base

Sample `main.ts` creating an AWS Network with [@envtio/base](https://www.npmjs.com/package/@envtio/base)

```console
synth -file main.ts -deps "@envtio/base:0.0.0" -src "cdktf.out/stacks/sample-stack"
```

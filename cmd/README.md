# synth CLI

Sample CLI using `go-synth` library with [bun.sh](https://bun.sh) executor.

```console
go build -o synth
./synth -file example/network.ts -deps "@envtio/base:0.0.0" -src "cdktf.out/stacks/network-stack" -out result/network
```

Will create directory `result/` containing all the synthesized files.

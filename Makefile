# TODO: Support other platforms
# TODO: use `-baseline` if AXV2 isn't supported
synth/bun:
	rm -rf bin
	mkdir -p bin
	curl -fsSLo bin/bun.zip https://github.com/oven-sh/bun/releases/download/bun-v1.1.23/bun-linux-x64.zip
	unzip bin/bun.zip -d bin
	mv bin/bun-linux-x64/bun bin/
	memexec-gen -dest synth/bun bin/bun
	rm -rf bin

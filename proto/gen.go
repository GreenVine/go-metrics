package proto

//go:generate go run github.com/bufbuild/buf/cmd/buf@v1.55.1 generate --clean
//go:generate go run github.com/bufbuild/buf/cmd/buf@v1.55.1 build ./ -o gen/descriptors.binpb

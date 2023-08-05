package pb

//go:generate protoc --go_out=paths=source_relative:. --go-drpc_out=paths=source_relative:. ./igloo.proto
//go:generate protoc --go_out=paths=source_relative:/data/Dev/blizzard/blizzard/pb --go-drpc_out=paths=source_relative:/data/Dev/blizzard/blizzard/pb ./igloo.proto

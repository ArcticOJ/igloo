//go:build !linux

package main

func main() {
	panic("igloo only works on Linux, other OSes are not supported and never will be, try using WSL (on Windows) or Docker (on MacOS).")
}

package main

import (
	"embed"

	"github.com/sweepyoface/conspire/cmd"
)

//go:embed static/*
var EmbedFs embed.FS

func main() {
	cmd.EmbedFs = EmbedFs
	cmd.Execute()
}

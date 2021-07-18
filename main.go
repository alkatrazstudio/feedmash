// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package main

import (
	_ "embed"
	App "feedmash/src"
)

//go:embed config.yaml
var exampleYaml string

func main() {
	App.Main(exampleYaml)
}

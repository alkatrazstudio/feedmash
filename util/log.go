// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package util

import (
	"fmt"
	"os"
)

func LogInfo(s interface{}) {
	_, _ = fmt.Fprintln(os.Stdout, s)
}

func LogWarn(s interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, s)
}

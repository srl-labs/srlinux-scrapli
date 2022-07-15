// Copyright 2021 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package srlinux

const (
	// PTY dimensions to override scrapli defaults,
	// since width=256 results in too long `---` delimiters in show commands
	defaultPtyWidth  = 140
	defaultPtyHeight = 60
)

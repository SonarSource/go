// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"fmt"
	"io"

	"golang.org/x/crypto/md4"
)

func main() {
	h := md4.New()
	data := "These pretzels are making me thirsty."
	io.WriteString(h, data)
	fmt.Printf("%x", h.Sum(nil))
}

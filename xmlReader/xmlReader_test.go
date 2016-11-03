// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlReader

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

// Test that a list l is not modified when calling MoveAfter or MoveBefore with a mark that is not an element of l.
func TestAll(t *testing.T) {
	str := `<?xml version="1.0" encoding="UTF-8"?>
<config>
	<common> 
	</common>
	<c1> 
		<old_pp a="1" b="2" c="3">			 
				<input_path>  a	</input_path> 
				<delete_same_file>N</delete_same_file> 
				<bak_path></bak_path> 
				<compress_flag></compress_flag> 
				cao
		</old_pp> 
		<new_pp>
				<input_path></input_path> 
				<delete_same_file>N</delete_same_file> 
				<bak_path></bak_path> 
				<compress_flag></compress_flag> 
		</new_pp> 
		<new_pp>
				<input_path></input_path> 
				<delete_same_file>N</delete_same_file> 
				<bak_path></bak_path> 
				<compress_flag></compress_flag> 
		</new_pp>
	</c1>
</config>`
	var d1 = []byte(str)
	err2 := ioutil.WriteFile("ppDiff.xml", d1, 0666) //写入文件(字节数组)
	if err2 != nil {
		fmt.Println("write file err")
	}
	defer os.Remove("ppDiff.xml")
	p, _ := NewXmlReader("ppDiff.xml")
	ret, _ := p.GetValue("/config/c1/old_pp")
	fmt.Println(ret, len(ret))
	names, _ := p.GetSubNodeName("/config/c1/old_pp/input_path")
	if names == nil {
		fmt.Println("nil")
	}
	fmt.Println(names)
	attr, _ := p.GetAttrMap("/config/c1/old_pp")
	fmt.Println(attr)
	return

}

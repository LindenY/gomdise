package gomdies

import (
	"fmt"
	"bytes"
)

type Args []interface{}

type Command struct {
	name string
	args Args
}

type Script struct {
	cmds []Command
	e error
}


func (c *Command) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(c.name)

	for str, i := range c.args {
		buffer.WriteString(" ")
		buffer.WriteString(str)
	}
	return buffer.String()
}
package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type Message struct {
	timestamp  time.Time
	sender     string
	remoteAddr net.Addr
	content    []byte
}

func (m *Message) Write(w io.WriteCloser) {
	message := fmt.Sprintf("%s %s %s :  %s",
		m.remoteAddr.String(),
		m.timestamp.Format("15:04:05"),
		m.sender,
		m.content,
	)

	w.Write([]byte(message))
}

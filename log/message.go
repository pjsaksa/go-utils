package log

import (
	"io"
	"time"
)

type Message struct {
	panic bool
	Line  string
}

func (m Message) write() {
	// Use forcedTime or the clock.
	var now time.Time
	if !forcedTime.IsZero() {
		now = forcedTime
	} else {
		now = time.Now()
	}

	line := now.Format("15:04:05.000") + " " + m.Line + "\n"
	io.WriteString(mainOutput, line)

	if m.panic {
		panic(m.Line)
	}
}

func LOG(msgList ...Message) {
	if msgList == nil {
		ERROR("nil")
	} else if len(msgList) == 1 {
		msgList[0].write()
	} else {
		var combined Message
		for idx, m := range msgList {
			if idx > 0 {
				combined.Line += " -> "
			}

			combined.Line += m.Line

			if m.panic {
				combined.panic = true
			}
		}
		combined.write()
	}
}

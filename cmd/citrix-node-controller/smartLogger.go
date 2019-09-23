package main

import (
	"fmt"
)

var (
	ConfigMapAdd    = 0x00000000
	ConfigMapDelete = 0x00000001
	ConfigMapUpdate = 0x00000002
)

type LogMessageQueue struct {
	mesages []string
}

func (queue *LogMessageQueue) Enqueue(message string) {
	queue.mesages = append(queue.mesages, message)
}

func (queue *LogMessageQueue) IsEmpty() bool {
	return len(queue.mesages) == 0
}

func (queue *LogMessageQueue) Dequeue() string {
	message := queue.mesages[0]
	queue.mesages = queue.mesages[1:len(queue.mesages)]
	return message
}

func (queue *LogMessageQueue) DumpMessage() {
	for queue.IsEmpty() != true {
		message := queue.Dequeue()
		fmt.Println(message)
	}
}

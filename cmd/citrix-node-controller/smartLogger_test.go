
package main

import (
	"testing"
)


func InitQueue() (*LogMessageQueue, *LogMessageQueue, *LogMessageQueue){

	ConfigMapAddEventQueueTest := LogMessageQueue{} 
	ConfigMapDeleteEventQueueTest := LogMessageQueue{}
	ConfigMapUpdateQueueTest  := LogMessageQueue{}
	return &ConfigMapAddEventQueueTest, &ConfigMapDeleteEventQueueTest, &ConfigMapUpdateQueueTest
}


func TestQueue(t *testing.T){
	AddEventQueue, DeleteEventQueue, UpdateEventQueue := InitQueue()
	AddEventQueue.Enqueue("[ADD] First Message ")        	
		if (AddEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty AddEventQueue")
		}
	AddEventQueue.Enqueue("[ADD] Second Message ")        	
		if (AddEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty AddEventQueue")
		}
	AddEventQueue.Enqueue("[ADD] Third  Message ")        	
		if (AddEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty AddEventQueue")
		}
	DeleteEventQueue.Enqueue("[DELETE] First Message ")        	
		if (DeleteEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty DeleteEventQueue")
		}
	DeleteEventQueue.Enqueue("[DELETE] Second Message ")        	
		if (DeleteEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty DeleteEventQueue")
		}
	DeleteEventQueue.Enqueue("[DELETE] Third  Message ")        
		if (DeleteEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty DeleteEventQueue")
		}
	UpdateEventQueue.Enqueue("[UPDATE] First Message ")        	
		if (UpdateEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty UpdateEventQueue")
		}
	UpdateEventQueue.Enqueue("[UPDATE] Second Message ")        	
		if (UpdateEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty UpdateEventQueue")
		}
	UpdateEventQueue.Enqueue("[UPDATE] Third  Message ")        
		if (UpdateEventQueue.IsEmpty() == true) {
			t.Error("Expected  NON Empty UpdateEventQueue")
		}
	DeleteEventQueue.DumpMessage()	
		if (DeleteEventQueue.IsEmpty() != true) {
			t.Error("Expected  Empty DeleteEventQueue")
		}
	UpdateEventQueue.DumpMessage()	
		if (UpdateEventQueue.IsEmpty() != true) {
			t.Error("Expected  Empty UpdateEventQueue")
		}
	AddEventQueue.DumpMessage()	
		if (AddEventQueue.IsEmpty() != true) {
			t.Error("Expected  Empty AddEventQueue")
		}
}

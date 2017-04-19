package main

import (
	"sync"
)

type Task interface {
	Start(wg *sync.WaitGroup)
	Shutdown()
}

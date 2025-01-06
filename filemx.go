package main

import "sync"

type FileMutex struct {
	mutexes sync.Map
}

func (f *FileMutex) Lock(path string) {
	mu, _ := f.mutexes.LoadOrStore(path, &sync.Mutex{})

	mu.(*sync.Mutex).Lock()
}

func (f *FileMutex) Unlock(path string) {
	mu, ok := f.mutexes.LoadAndDelete(path)
	if !ok {
		return
	}

	mu.(*sync.Mutex).Unlock()
}

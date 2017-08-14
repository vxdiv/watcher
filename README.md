[![Build Status](https://travis-ci.org/vxdiv/watcher.svg?branch=master)](https://travis-ci.org/vxdiv/watcher)
[![codecov](https://codecov.io/gh/vxdiv/watcher/branch/master/graph/badge.svg)](https://codecov.io/gh/vxdiv/watcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/vxdiv/watcher)](https://goreportcard.com/report/github.com/vxdiv/watcher)

# About

This package provides the ability to run background tasks that are performed at the regular intervals. 

## Install

`go get github.com/vxdiv/watcher`

## Usage

### Simple job
```go
package main

import (
	"fmt"
	"time"

	"github.com/vxdiv/watcher"
)

func main() {

	// Start watcher
	watcher.Run()

	// Start Job in background with interval 3 sec
	watcher.Start(func() {
		fmt.Println("working")
	}, time.Second*3, watcher.CompositeKey{111, "test_1"})

	time.Sleep(time.Second * 10)
	fmt.Println("Done")
}

```

### You can stop job inside job itself or in other place your code even in other package

```go
package main

import (
	"fmt"
	"time"

	"github.com/vxdiv/watcher"
)

func main() {

	watcher.Run()

	watcher.Start(job(0), time.Second*1, watcher.CompositeKey{111, "test_1"})

	time.Sleep(time.Second * 10)
	fmt.Println("Done")
}

func job(c int) watcher.JobFunc {
	count := c
	return func() {
		count++
		fmt.Println(count)
		if count == 5 {
			fmt.Println("job stop")
			watcher.Stop(watcher.CompositeKey{111, "test_1"})
		}
	}
}

```



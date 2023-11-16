package util

import (
	"fmt"
	"time"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start).Milliseconds()
	fmt.Printf("%v took %v ms \n", name, elapsed)
}

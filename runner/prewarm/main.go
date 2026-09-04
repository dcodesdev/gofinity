// Package main exists only to fill the build cache baked into the runner
// image. It imports the standard library packages a Go challenge is likely to
// reach for, so the first real run in a container compiles almost nothing.
//
// It is never executed by a submission.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done() }()
	wg.Wait()

	values := []int{3, 1, 2}
	sort.Ints(values)
	slices.Sort(values)
	m := map[string]int{"a": 1}
	_ = slices.Collect(maps.Keys(m))

	encoded, _ := json.Marshal(values)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	fmt.Fprint(writer, strings.TrimSpace(string(encoded)),
		strconv.Itoa(int(math.Abs(-1))),
		time.Now().IsZero(),
		errors.New("unused"))
}

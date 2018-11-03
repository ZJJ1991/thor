package main

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	consoleFlages = []cli.Flag{}
)

// func producer(chnl chan int) {
// 	for i := 0; i < 10; i++ {
// 		chnl <- i
// 	}
// 	close(chnl)
// }
// func main() {
// 	ch := make(chan int)
// 	go producer(ch)
// 	for {
// 		v, ok := <-ch
// 		if ok == false {
// 			break
// 		}
// 		fmt.Println("Received ", v, ok)
// 	}
// }
// func main() {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	go func() {
// 		exitSignalCh := make(chan os.Signal)
// 		signal.Notify(exitSignalCh, os.Interrupt, os.Kill, syscall.SIGTERM)

// 		select {
// 		case sig := <-exitSignalCh:
// 			fmt.Println("exit signal received", "signal", sig)
// 			cancel()
// 		}
// 	}()
// 	fmt.Println("exit signal received", ctx)
// 	// Pass a context with a timeout to tell a blocking function that it
// 	// should abandon its work after the timeout elapses.
// }

// SafeCounter is safe to use concurrently.
// type SafeCounter struct {
// 	v   map[string]int
// 	txt map[string]int
// 	mux sync.Mutex
// }

// // Inc increments the counter for the given key.
// func (c *SafeCounter) Inc(key string) {
// 	c.mux.Lock()
// 	// Lock so only one goroutine at a time can access the map c.v.
// 	c.v[key]++
// 	fmt.Println(c.v[key])
// 	c.mux.Unlock()
// }

// // Value returns the current value of the counter for the given key.
// func (c *SafeCounter) Value(key string) int {
// 	c.mux.Lock()
// 	// Lock so only one goroutine at a time can access the map c.v.
// 	defer c.mux.Unlock()
// 	return c.v[key]
// }

// func (c *SafeCounter) test(key string) {
// 	c.mux.Lock()
// 	fmt.Println(key)
// 	c.txt[key] = 88
// 	c.mux.Unlock()
// }

// func (c *SafeCounter) result(key string) int {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	return c.txt[key]
// }
// func main() {
// 	c := SafeCounter{v: make(map[string]int),
// 		txt: make(map[string]int)}
// 	for i := 0; i < 1111; i++ {
// 		go c.Inc("somekey")
// 	}
// 	time.Sleep(time.Second)
// 	c.test("txtkey")
// 	fmt.Println(c.result("txtkey"))
// 	fmt.Println("try to print")
// 	fmt.Println(c.Value("somekey"))
// }

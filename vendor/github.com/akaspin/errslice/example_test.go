package errslice_test

import (
	"fmt"
	"github.com/akaspin/errslice"
)

func Example_loop() {
	var err error
	for i := 0; i < 5; i++ {
		if i%2 == 0 {
			err = errslice.Append(err, fmt.Errorf("bad:%d", i))
		}
	}
	fmt.Println(err)
	// Output:
	// bad:0,bad:2,bad:4
}

func Example_assert() {
	var err error
	for i := 0; i < 5; i++ {
		if i%2 == 0 {
			err = errslice.Append(err, fmt.Errorf("bad:%d", i))
		}
	}
	switch err1 := err.(type) {
	case errslice.Error:
		for _, err2 := range err1 {
			fmt.Println(err2)
		}
	case nil:
		fmt.Println("all good")
	default:
		fmt.Println(err1)
	}
	// Output:
	// bad:0
	// bad:2
	// bad:4
}

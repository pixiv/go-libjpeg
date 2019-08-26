package jpeg_test

import (
	"bytes"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/pixiv/go-libjpeg/jpeg"
)

// https://github.com/pixiv/go-libjpeg/issues/58
func TestDecode58(t *testing.T) {
	data, err := ioutil.ReadFile("../test/images/63596186-9075ca80-c5b2-11e9-92d0-94aca72538c3.jpeg")
	if err != nil {
		t.Fatalf("Error: %v\n", err)
		return
	}

	const PARALLEL = 16
	wait := &sync.WaitGroup{}
	for i := 0; i < PARALLEL; i++ {
		go func() {
			wait.Add(1)
			defer wait.Done()
			processImage(t, data, 100)
		}()
	}

	wait.Wait()
}

func processImage(t *testing.T, data []byte, times int) {
	for i := 0; i < times; i++ {
		_, err := jpeg.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			t.Errorf("expected decode error, but not")
		}
	}
}

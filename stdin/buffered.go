package stdin

import (
	"bufio"
	"context"
	"os"
)

type BufferedReader struct {
	bufferSize int
}

func NewBufferedReader(bufferSize int) *BufferedReader {
	r := BufferedReader{
		bufferSize: bufferSize,
	}

	return &r
}

func (r *BufferedReader) Read(ctx context.Context, notifyError func(err error)) chan string {
	out := make(chan string, r.bufferSize)

	go func() {
		defer close(out)

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			input := scanner.Text()
			select {
			case <-ctx.Done():
				return
			case out <- input:
			}
		}

		if err := scanner.Err(); err != nil {
			notifyError(err)
		}
	}()

	return out
}

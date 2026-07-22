package ledger

import "context"

// Merge forwards values from two inputs until both close or the context ends.
func Merge[T any](ctx context.Context, left, right <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for left != nil || right != nil {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-left:
				if !ok {
					left = nil
					continue
				}
				select {
				case out <- value:
				case <-ctx.Done():
					return
				}
			case value, ok := <-right:
				if !ok {
					right = nil
					continue
				}
				select {
				case out <- value:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

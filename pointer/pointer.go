package pointer

// Cursor represents a cursor over a collection of type T
type Cursor[T any] struct {
	position   int
	collection []T
}

// NewCursor creates a new cursor for the given collection
func NewCursor[T any](collection []T) *Cursor[T] {
	return &Cursor[T]{
		position:   0,
		collection: collection,
	}
}

// Current returns the current element in the collection
func (c *Cursor[T]) Current() T {
	return c.collection[c.position]
}

// Consume returns the current element and moves the cursor forward
func (c *Cursor[T]) Consume() T {
	current := c.Current()
	c.position++
	return current
}

// IsOpen returns true if the cursor hasn't reached the end of the collection
func (c *Cursor[T]) IsOpen() bool {
	return c.position < len(c.collection)
}

package otelexample

const (
	// DefaultPageSize is the elements count that will be placed
	// on the single response page if option was not specified.
	DefaultPageSize = 20

	// MaxPageSize is the maximum elements count that
	// could be placed on the single response page.
	MaxPageSize = 100
)

// FindOptions represents options passed to all find methods
// with multiple results.
type FindOptions struct {
	limit  uint64
	offset uint64
}

// NewFindOptions returns a new FindOptions instance.
func NewFindOptions(limit, offset uint64) FindOptions {
	opts := FindOptions{
		limit:  limit,
		offset: offset,
	}

	if opts.limit == 0 {
		opts.limit = DefaultPageSize
	}

	if opts.limit > MaxPageSize {
		opts.limit = MaxPageSize
	}

	return opts
}

// Limit is the elements count that could be placed on the single
// response page.
func (opts FindOptions) Limit() uint64 {
	return opts.limit
}

// Offset is the elements count that should be skipped.
func (opts FindOptions) Offset() uint64 {
	return opts.offset
}

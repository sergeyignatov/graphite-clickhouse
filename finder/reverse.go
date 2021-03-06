package finder

import (
	"bytes"
	"context"
	"strings"
	"time"
)

type ReverseFinder struct {
	wrapped    Finder
	baseFinder Finder
	url        string        // clickhouse dsn
	table      string        // graphite_reverse_tree table
	timeout    time.Duration // clickhouse query timeout
	isUsed     bool          // use reverse table
}

func ReverseString(target string) string {
	a := strings.Split(target, ".")
	l := len(a)
	for i := 0; i < l/2; i++ {
		a[i], a[l-i-1] = a[l-i-1], a[i]
	}

	return strings.Join(a, ".")
}

func ReverseBytes(target []byte) []byte {
	// @TODO: check performance
	a := bytes.Split(target, []byte{'.'})

	l := len(a)
	for i := 0; i < l/2; i++ {
		a[i], a[l-i-1] = a[l-i-1], a[i]
	}

	return bytes.Join(a, []byte{'.'})
}

func WrapReverse(f Finder, url string, table string, timeout time.Duration) *ReverseFinder {
	return &ReverseFinder{
		wrapped:    f,
		baseFinder: NewBase(url, table, timeout),
		url:        url,
		table:      table,
		timeout:    timeout,
	}
}

func (r *ReverseFinder) Execute(ctx context.Context, query string, from int64, until int64) error {
	p := strings.LastIndexByte(query, '.')
	if p < 0 || p >= len(query)-1 {
		return r.wrapped.Execute(ctx, query, from, until)
	}

	if HasWildcard(query[p+1:]) {
		return r.wrapped.Execute(ctx, query, from, until)
	}

	r.isUsed = true
	return r.baseFinder.Execute(ctx, ReverseString(query), from, until)
}

func (r *ReverseFinder) List() [][]byte {
	if !r.isUsed {
		return r.wrapped.List()
	}

	list := r.baseFinder.List()
	for i := 0; i < len(list); i++ {
		list[i] = ReverseBytes(list[i])
	}

	return list
}

func (r *ReverseFinder) Series() [][]byte {
	if !r.isUsed {
		return r.wrapped.Series()
	}

	list := r.baseFinder.Series()
	for i := 0; i < len(list); i++ {
		list[i] = ReverseBytes(list[i])
	}

	return list
}

func (r *ReverseFinder) Abs(v []byte) []byte {
	return v
}

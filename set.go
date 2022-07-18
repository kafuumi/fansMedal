package fans

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(items ...T) {
	for i := range items {
		s[items[i]] = struct{}{}
	}
}

func (s Set[T]) Contains(item T) bool {
	_, ok := s[item]
	return ok
}

func (s Set[T]) Remove(item ...T) {
	for i := range item {
		delete(s, item[i])
	}
}

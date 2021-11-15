package rip

type NotFoundError struct {
	Resource string
}

func (e NotFoundError) Error() string {
	return "resource not found: " + e.Resource
}

// error.go

package models

type C2CError struct {
	ErrorCode string
}

func (c *C2CError) Error() string {
	return c.ErrorCode
}

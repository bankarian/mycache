package core

// ByteView is read-only view of bytes, who implements Value
type ByteView struct {
	bs []byte // bs stores the real cache content in any type
}

// Len returns the view's length
func (v ByteView) Len() int {
	return len(v.bs)
}

func (v *ByteView) ByteSlice() []byte {
	return clone(v.bs)
}

func (v *ByteView) String() string {
	return string(v.bs)
}

func clone(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

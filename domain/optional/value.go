package optional

func BoolValue(in bool) *Bool {
	out := Bool(in)

	return &out
}

type Bool bool

func (v *Bool) IsTrue() bool {
	return v != nil && bool(*v)
}

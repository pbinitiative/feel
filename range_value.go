package feel

import (
// "fmt"
)

type RangeValue struct {
	StartOpen bool
	Start     any

	EndOpen bool
	End     any
}

func (self RangeValue) BeforePoint(p any) (bool, error) {
	pos, err := self.Position(p)
	if err != nil {
		return false, err
	}
	return pos > 0, nil

}

func (self RangeValue) AfterPoint(p any) (bool, error) {
	pos, err := self.Position(p)
	if err != nil {
		return false, err
	}
	return pos < 0, nil
}

func (self RangeValue) BeforeRange(other RangeValue) (bool, error) {
	r, err := compareInterfaces(self.End, other.Start)
	if err != nil {
		return false, err
	}

	if !self.EndOpen && !other.StartOpen {
		// two ranges meet
		return r < 0, nil
	} else {
		return r <= 0, nil
	}
}

func (self RangeValue) AfterRange(other RangeValue) (bool, error) {
	r, err := compareInterfaces(self.Start, other.End)
	if err != nil {
		return false, err
	}
	if !self.StartOpen && !other.EndOpen {
		return r > 0, nil
	} else {
		return r >= 0, nil
	}
}

func (self RangeValue) Position(p any) (int, error) {
	cmpStart, err := compareInterfaces(p, self.Start)
	if err != nil {
		return 0, err
	}
	if self.StartOpen {
		if cmpStart <= 0 {
			return -1, nil
		}
	} else {
		if cmpStart == 0 {
			return 0, nil
		} else if cmpStart < 0 {
			return -1, nil
		}
	}

	cmpEnd, err := compareInterfaces(p, self.End)
	if err != nil {
		return 0, err
	}
	if self.EndOpen && cmpEnd >= 0 {
		return 1, nil
	} else if !self.EndOpen && cmpEnd > 0 {
		return 1, nil
	}
	return 0, nil

}

func (self RangeValue) Includes(other RangeValue) (bool, error) {
	cmpStart, err := compareInterfaces(self.Start, other.Start)
	if err != nil {
		return false, err
	}
	cmpEnd, err := compareInterfaces(self.End, other.End)
	if err != nil {
		return false, err
	}
	if cmpStart > 0 || cmpEnd < 0 {
		return false, nil
	}

	if !(cmpStart < 0 || !self.StartOpen || other.StartOpen) {
		return false, nil
	}

	if !(cmpEnd > 0 || !self.EndOpen || other.EndOpen) {
		return false, nil
	}
	return true, nil
}

func (self RangeValue) Contains(p any) bool {
	r, err := self.Position(p)
	if err != nil {
		panic(err)
	}
	return r == 0
}

func (self RangeValue) overlapsBefore(other RangeValue) (bool, error) {
	pos, err := other.Position(self.End)
	if err != nil {
		return false, err
	}
	if pos != 0 {
		return false, nil
	} else if self.EndOpen && CompareValues(self.End, other.Start) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (self RangeValue) overlapsAfter(other RangeValue) (bool, error) {
	pos, err := other.Position(self.Start)
	if err != nil {
		return false, err
	}
	if pos != 0 {
		return false, nil
	} else if self.EndOpen && CompareValues(self.Start, other.End) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func installRangeFunctions(prelude *Prelude) {
	prelude.Bind("before", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		a := kwargs["a"]
		b := kwargs["b"]
		switch va := a.(type) {
		case *RangeValue:
			switch vb := b.(type) {
			case *RangeValue:
				return va.BeforeRange(*vb)
			default:
				return va.BeforePoint(vb)
			}
		default:
			switch vb := b.(type) {
			case *RangeValue:
				return vb.AfterPoint(va)
			default:
				cmp, err := compareInterfaces(va, vb)
				if err != nil {
					return nil, err
				} else {
					return cmp < 0, nil
				}
			}
		}
	}).Required("a", "b"))

	prelude.Bind("after", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		a := kwargs["a"]
		b := kwargs["b"]
		switch va := a.(type) {
		case *RangeValue:
			switch vb := b.(type) {
			case *RangeValue:
				return va.AfterRange(*vb)
			default:
				return va.AfterPoint(vb)
			}
		default:
			switch vb := b.(type) {
			case *RangeValue:
				return vb.BeforePoint(va)
			default:
				cmp, err := compareInterfaces(va, vb)
				if err != nil {
					return nil, err
				} else {
					return cmp > 0, nil
				}
			}
		}
	}).Required("a", "b"))

	prelude.Bind("meets", wrapTyped(func(a *RangeValue, b *RangeValue) (bool, error) {
		r, err := compareInterfaces(a.End, b.Start)
		if err != nil {
			return false, err
		}
		return !a.EndOpen && !b.StartOpen && r == 0, nil
	}).Required("a", "b"))

	prelude.Bind("met by", wrapTyped(func(a *RangeValue, b *RangeValue) (bool, error) {
		r, err := compareInterfaces(a.Start, b.End)
		if err != nil {
			return false, err
		}
		return !b.EndOpen && !a.StartOpen && r == 0, nil
	}).Required("a", "b"))

	prelude.Bind("overlaps", wrapTyped(func(a *RangeValue, b *RangeValue) (bool, error) {
		obefore, err := a.overlapsBefore(*b)
		if err != nil {
			return false, err
		} else if obefore {
			return true, nil
		}
		oafter, err := a.overlapsAfter(*b)
		if err != nil {
			return false, err
		} else if oafter {
			return true, nil
		}
		return false, nil
	}).Required("a", "b"))

	prelude.Bind("overlaps before", wrapTyped(func(a *RangeValue, b *RangeValue) (bool, error) {
		return a.overlapsBefore(*b)
	}).Required("a", "b"))

	prelude.Bind("overlaps after", wrapTyped(func(a *RangeValue, b *RangeValue) (bool, error) {
		return a.overlapsAfter(*b)
	}).Required("a", "b"))

	prelude.Bind("finishes", wrapTyped(func(a any, b *RangeValue) (bool, error) {
		switch va := a.(type) {
		case *RangeValue:
			r, err := compareInterfaces(va.End, b.End)
			if err != nil {
				return false, err
			}
			return r == 0 && !b.EndOpen, nil
		default:
			r, err := compareInterfaces(a, b.End)
			if err != nil {
				return false, err
			}
			return r == 0 && !b.EndOpen, nil
		}
	}).Required("a", "b"))

	prelude.Bind("starts", wrapTyped(func(a any, b *RangeValue) (bool, error) {
		switch va := a.(type) {
		case *RangeValue:
			r, err := compareInterfaces(va.Start, b.Start)
			if err != nil {
				return false, err
			}
			if r != 0 {
				return false, nil
			}
			if va.StartOpen && !b.StartOpen {
				return false, nil
			}
			if !va.StartOpen && b.StartOpen {
				return false, nil
			}
			return true, nil
		default:
			r, err := compareInterfaces(a, b.Start)
			if err != nil {
				return false, err
			}
			return r == 0 && !b.StartOpen, nil
		}
	}).Required("a", "b"))

	prelude.Bind("finished by", wrapTyped(func(a *RangeValue, b any) (bool, error) {
		switch vb := b.(type) {
		case *RangeValue:
			r, err := compareInterfaces(vb.End, a.End)
			if err != nil {
				return false, err
			}

			return r == 0 && !a.EndOpen, nil
		default:
			r, err := compareInterfaces(b, a.End)
			if err != nil {
				return false, err
			}
			return r == 0 && !a.EndOpen, nil
		}
	}).Required("a", "b"))

	prelude.Bind("started by", wrapTyped(func(a *RangeValue, b any) (bool, error) {
		switch vb := b.(type) {
		case *RangeValue:
			r, err := compareInterfaces(vb.Start, a.Start)
			if err != nil {
				return false, err
			}
			if r != 0 {
				return false, nil
			}
			if vb.StartOpen && !a.StartOpen {
				return false, nil
			}
			if !vb.StartOpen && a.StartOpen {
				return false, nil
			}
			return true, nil
		default:
			r, err := compareInterfaces(b, a.Start)
			if err != nil {
				return false, err
			}
			return r == 0 && !a.StartOpen, nil
		}
	}).Required("a", "b"))

	prelude.Bind("includes", wrapTyped(func(a *RangeValue, b any) (bool, error) {
		switch vb := b.(type) {
		case *RangeValue:
			return a.Includes(*vb)
		default:
			r, err := a.Position(vb)
			if err != nil {
				return false, err
			}
			return r == 0, nil
		}
	}).Required("a", "b"))

	prelude.Bind("during", wrapTyped(func(a any, b *RangeValue) (bool, error) {
		switch va := a.(type) {
		case *RangeValue:
			return b.Includes(*va)
		default:
			r, err := b.Position(va)
			if err != nil {
				return false, err
			}
			return r == 0, nil
		}
	}).Required("a", "b"))

}

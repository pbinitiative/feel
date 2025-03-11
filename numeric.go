package feel

// spec on FEEL's number is https://kiegroup.github.io/dmn-feel-handbook/#number
import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"
)

const (
	Prec = 34 * 8
)

var (
	ErrParseNumber = errors.New("fail to parse number")
)

type Number struct {
	v *big.Float
}

func NewNumber(strn string) *Number {
	v := new(big.Float)
	v.SetPrec(Prec).SetString(strn)
	return &Number{v: v}
}

func NewNumberFromInt64(input int64) *Number {
	v := new(big.Float)
	v.SetPrec(200).SetInt64(input)
	return &Number{v: v}
}

func NewNumberFromFloat(input float64) *Number {
	v := new(big.Float)
	v.SetPrec(200).SetFloat64(input)
	return &Number{v: v}
}

func ParseNumberWithErr(v interface{}) (*Number, error) {
	switch vv := v.(type) {
	case int:
		return NewNumberFromInt64(int64(vv)), nil
	case int64:
		return NewNumberFromInt64(vv), nil
	case float64:
		return NewNumberFromFloat(vv), nil
	case string:
		return NewNumber(vv), nil
	case *Number:
		return vv, nil
	default:
		return nil, ErrParseNumber
	}
}

func N(v interface{}) *Number {
	n, err := ParseNumberWithErr(v)
	if err != nil {
		panic(err)
	}
	return n
}

func (n Number) Int64() int64 {
	i64v, _ := n.v.Int64()
	return i64v
}

func (n Number) Int() int {
	return int(n.Int64())
}

func (n Number) Float64() float64 {
	f64v, _ := n.v.Float64()
	return f64v
}

func (n *Number) Add(other *Number) *Number {
	newv := new(big.Float)
	newv.SetPrec(Prec).Add(n.v, other.v)
	return &Number{v: newv}
}

func (n *Number) Sub(other *Number) *Number {
	newv := new(big.Float)
	newv.SetPrec(Prec).Sub(n.v, other.v)
	return &Number{v: newv}
}

func (n *Number) Mul(other *Number) *Number {
	newv := new(big.Float)
	newv.SetPrec(Prec).Mul(n.v, other.v)
	return &Number{v: newv}
}

func (n *Number) Cmp(other *Number) int {
	return n.v.Cmp(other.v)
}

func (n *Number) IntDiv(other *Number) *Number {
	newv := new(big.Int)
	a, _ := n.v.Int(nil)
	b, _ := other.v.Int(nil)
	newv.Div(a, b)
	newf := new(big.Float)
	newf.SetPrec(Prec).SetInt(newv)
	return &Number{v: newf}
}

func (n *Number) FloatDiv(other *Number) *Number {
	a, _ := n.v.Float64()
	b, _ := other.v.Float64()
	newf := new(big.Float)
	newf.SetPrec(Prec).SetFloat64(a / b)
	return &Number{v: newf}
}

func (n *Number) IntMod(other *Number) *Number {
	newv := new(big.Int)
	a, _ := n.v.Int(nil)
	b, _ := other.v.Int(nil)
	newv.Mod(a, b)
	newf := new(big.Float)
	newf.SetPrec(Prec).SetInt(newv)
	return &Number{v: newf}
}

func (n Number) Equal(other Number) bool {
	return n.Compare(other) == 0
}

func (n Number) Compare(other Number) int {
	return n.v.Cmp(other.v)
}

func (n Number) String() string {
	s := n.v.Text('f', 18)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func (n Number) MarshalJSON() ([]byte, error) {
	if f32v, acc := n.v.Float32(); acc == big.Exact {
		return json.Marshal(f32v)
	}
	return json.Marshal(n.String())
}

var Zero = N(0)

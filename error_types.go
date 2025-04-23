package feel

import (
	"fmt"
	"strings"
)

type EvalError struct {
	Code    int
	Short   string
	Message string
}

func (evalError EvalError) Error() string {
	return fmt.Sprintf("%d %s, %s", evalError.Code, evalError.Short, evalError.Message)
}

func NewEvalError(code int, short string, msgs ...string) *EvalError {
	message := strings.Join(msgs, " ")
	return &EvalError{
		Code:    code,
		Short:   short,
		Message: message,
	}
}

func NewErrKeyNotFound(keyName string) *EvalError {
	return NewEvalError(-4000, "key not found", fmt.Sprintf("cannot get key '%s'", keyName))
}

func NewErrIndex(msg string) *EvalError {
	return NewEvalError(-4001, "index error", msg)
}

func NewErrTypeMismatch(expectType string) *EvalError {
	return NewEvalError(-4002, "type mismatch", "expect", expectType)
}

// NewErrKeywordArgument argument errors
func NewErrKeywordArgument(argName string) *EvalError {
	return NewEvalError(-4010, "keyword argument required", "require keyword arg", argName)
}

func NewErrTooFewArguments(required []string) *EvalError {
	reqArgs := strings.Join(required, ", ")
	return NewEvalError(-4011, "too few argument", "require arguments:", reqArgs)
}

func NewErrTooManyArguments() *EvalError {
	return NewEvalError(-4012, "too many arguments")
}

func NewErrBadOp(leftType, op, rightType string) *EvalError {
	return NewEvalError(-5001, "type mismatch in op", "bad types in op, ", leftType, op, rightType)
}

func NewErrBadInput(row int, column int, inputText string) *EvalError {
	return NewEvalError(-6001, "bad input", fmt.Sprintf(`at position row:%d, column:%d, input:%s`, row, column, inputText))
}

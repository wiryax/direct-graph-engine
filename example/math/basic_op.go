package math

type mathOp int

const (
	Add mathOp = iota
	Sub
	Mul
	Div
)

// func basicMath(rState *engine.RuntimeState, x string, y int, op mathOp) error {
// 	v, err := rState.GetVariable(x)
// 	if err != nil || v == nil {
// 		return fmt.Errorf("error while get variable with id %s", x)
// 	}
// 	i, err := strconv.Atoi(*v)
// 	if err != nil {
// 		return err
// 	}

// 	switch op {
// 	case Add:
// 		i += y
// 	case Sub:
// 		i -= y
// 	case Mul:
// 		i *= y
// 	case Div:
// 		i /= y
// 	default:
// 		return fmt.Errorf("unknown math operator")
// 	}

// 	result := strconv.Itoa(i)
// 	rState.SetVariable(x, &result)
// 	return nil
// }

// func Additional(x string, y int) func(rState *runtimeState) error {
// 	return func(rState *runtimeState) error {
// 		return basicMath(rState, x, y, Add)
// 	}
// }

// func Subtraction(x string, y int) func(rState *runtimeState) error {
// 	return func(rState *runtimeState) error {
// 		return basicMath(rState, x, y, Sub)
// 	}
// }

// func Division(x string, y int) func(rState *runtimeState) error {
// 	return func(rState *runtimeState) error {
// 		return basicMath(rState, x, y, Div)
// 	}
// }

// func Multiplication(x string, y int) func(rState *runtimeState) error {
// 	return func(rState *runtimeState) error {
// 		return basicMath(rState, x, y, Mul)
// 	}
// }

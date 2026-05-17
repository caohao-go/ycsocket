package types

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type CompOp uint8

const (
	OpNone CompOp = iota // 未知
	OpEq                 // =
	OpGt                 // >
	OpLt                 // <
)

// IsNumber 判断是否为数字
func IsNumber(str string) bool {
	if str == "" || str == "-" || str == "+" || str == "." {
		return false
	}

	var isPoint bool
	if str[0] == '.' {
		isPoint = true
	} else if !(str[0] == '-' || str[0] == '+' || (str[0] >= '0' && str[0] <= '9')) {
		return false
	}

	for _, ch := range str[1:] {
		if ch == '.' {
			if isPoint { // 两个小数点
				return false
			} else {
				isPoint = true
			}
		} else if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}

// CompareNumber 数字比较，比较 a、b 的大小，结果有 =、>、<
func CompareNumber(a, b interface{}) CompOp {
	op := quickCompare(a, b)
	if op != OpNone {
		return op
	}

	aIsDigit, aNegative, _, aIntegerPart, aDecimalPart := ValueInfo(a)
	bIsDigit, bNegative, _, bIntegerPart, bDecimalPart := ValueInfo(b)

	if !aIsDigit || !bIsDigit {
		aStr := ToString(a)
		bStr := ToString(b)
		if aStr == bStr {
			return OpEq
		} else if aStr > bStr {
			return OpGt
		} else {
			return OpLt
		}
	}

	if aNegative == bNegative && aIntegerPart == bIntegerPart && aDecimalPart.Equal(bDecimalPart) {
		return OpEq
	}

	if !aNegative { // a 为正数
		if bNegative { // b 为负数
			return OpGt
		} else { // a、b 都为正数
			if aIntegerPart > bIntegerPart {
				return OpGt
			} else if aIntegerPart < bIntegerPart {
				return OpLt
			} else { // a、b 整数部分相同
				if aDecimalPart.LessThan(bDecimalPart) {
					return OpLt
				} else {
					return OpGt
				}
			}
		}

	} else { // a 为负数
		if !bNegative { //b 为正数
			return OpLt
		} else { // a、b 都为负数
			if aIntegerPart > bIntegerPart {
				return OpLt
			} else if aIntegerPart < bIntegerPart {
				return OpGt
			} else { // a、b 整数部分相同
				if aDecimalPart.LessThan(bDecimalPart) {
					return OpGt
				} else {
					return OpLt
				}
			}
		}
	}
}

// IsEqual 判断两个值是否相等
func IsEqual(a, b interface{}) bool {
	switch a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return CompareNumber(a, b) == OpEq
	case bool:
		return ToBool(a) == ToBool(b)
	}

	switch b.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return CompareNumber(a, b) == OpEq
	case bool:
		return ToBool(a) == ToBool(b)
	}

	return ToString(a) == ToString(b)
}

// 快速比较
func quickCompare(a, b interface{}) CompOp {
	switch a.(type) {
	case int, int8, int16, int32, int64:
		switch b.(type) {
		case int, int8, int16, int32, int64:
			av, _ := ToInt64(a)
			bv, _ := ToInt64(b)
			if av == bv {
				return OpEq
			} else if av > bv {
				return OpGt
			} else {
				return OpLt
			}
		case uint, uint8, uint16, uint32, uint64:
			av, _ := ToInt64(a)
			if av < 0 {
				return OpLt
			}

			bv, _ := ToUint64(b)
			if uint64(av) == bv {
				return OpEq
			} else if uint64(av) > bv {
				return OpGt
			} else {
				return OpLt
			}
		}
	case uint, uint8, uint16, uint32, uint64:
		switch b.(type) {
		case uint, uint8, uint16, uint32, uint64:
			av, _ := ToUint64(a)
			bv, _ := ToUint64(b)
			if av == bv {
				return OpEq
			} else if av > bv {
				return OpGt
			} else {
				return OpLt
			}
		case int, int8, int16, int32, int64:
			bv, _ := ToInt64(b)
			if bv < 0 {
				return OpGt
			}
			av, _ := ToUint64(a)
			if av == uint64(bv) {
				return OpEq
			} else if av > uint64(bv) {
				return OpGt
			} else {
				return OpLt
			}
		}
	case float64:
		bv, ok := b.(float64)
		if ok {
			if a.(float64) == bv {
				return OpEq
			} else if a.(float64) > bv {
				return OpGt
			} else {
				return OpLt
			}
		}
	}

	return OpNone
}

// ValueInfo 获取值的类型
func ValueInfo(val interface{}) (isDigit, negative, isPoint bool, integerPart uint64, decimalPart decimal.Decimal) {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		isDigit = true
		iv, _ := ToInt64(v)
		if iv < 0 {
			negative = true
			iv = -iv
		}
		integerPart = uint64(iv)
		return
	case uint, uint8, uint16, uint32, uint64:
		isDigit = true
		integerPart, _ = ToUint64(v)
		return
	case float32:
		return floatValueInfo(float64(v))
	case float64:
		return floatValueInfo(v)
	}

	str := strings.TrimSpace(ToString(val))
	if val == "" || val == "-" || val == "+" || val == "." {
		return false, false, false, 0, decimal.Decimal{}
	}

	var pointIndex int

	if str[0] == '-' {
		negative = true
	} else if str[0] == '.' {
		isPoint = true
		pointIndex = 0
	} else if !(str[0] == '+' || (str[0] >= '0' && str[0] <= '9')) {
		return false, false, false, 0, decimal.Decimal{}
	}

	for k, ch := range str[1:] {
		if ch == '.' {
			if isPoint { // 两个小数点
				return false, false, false, 0, decimal.Decimal{}
			} else {
				isPoint = true
				pointIndex = k + 1
			}
		} else if ch < '0' || ch > '9' {
			return false, false, false, 0, decimal.Decimal{}
		}
	}

	isDigit = true

	if isPoint {
		integerPartStr := str[:pointIndex]
		decimalPartStr := fmt.Sprintf("0.%s", str[pointIndex+1:])

		if negative {
			integerPartStr = integerPartStr[1:]
		}

		integerPart, _ = ToUint64(integerPartStr)
		decimalPart, _ = decimal.NewFromString(decimalPartStr)
	} else {
		if negative {
			str = str[1:]
		}
		integerPart, _ = ToUint64(str)
	}

	return
}

func floatValueInfo(v float64) (isDigit, negative, isPoint bool, integerPart uint64, decimalPart decimal.Decimal) {
	isDigit = true
	isPoint = true

	if v < 0 {
		negative = true
		v = -v
	}

	integerPart = uint64(v)
	decimalPart = decimal.NewFromFloat(v).Sub(decimal.NewFromInt(int64(integerPart)))
	return
}

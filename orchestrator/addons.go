package main

import (
	"fmt"
	"strings"
)

func createStack(expression string) ([]string, bool) {
	res := false
	for index, char := range expression {
		fmt.Printf("Index: %d, Char: '%c'\n", index, char)
		if char == '+' {
			res = true
			break
		}
		if char == '-' {
			res = true
			break
		}
		if char == '*' {
			res = true
			break
		}
		if char == '/' {
			res = true
			break
		}
	}

	if !res {
		return nil, false
	}

	tokens := strings.Split(strings.ReplaceAll(expression, " ", ""), "")
	postfix, err := convertToPostfix(tokens)

	if !err {
		return nil, false
	}

	if len(postfix) < 3 {
		return postfix, false
	}

	return postfix, true
}

func precedence(op string) int {
	prio, _ := operations[op]
	return prio
}

func convertToPostfix(infix []string) ([]string, bool) {
	var output []string
	stack := make([]string, 0)
	for _, token := range infix {
		if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1] != "(" {
				return nil, false
			}
			stack = stack[:len(stack)-1] // удалить '('
		} else if isMathOperator(token) {
			for len(stack) > 0 && stack[len(stack)-1] != "(" && precedence(token) <= precedence(stack[len(stack)-1]) {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else { // число
			output = append(output, token)
		}
	}
	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, false
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}
	return output, true
}

func isMathOperator(s string) bool {
	_, ok := operations[s]
	return ok
}

var operations = map[string]int{
	"+": 1,
	"-": 1,
	"*": 2,
	"/": 2,
}

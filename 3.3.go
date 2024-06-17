package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)


//переменная cтэка
var memory memoryStack


//
type memoryLayer struct {
	variables map[string]varData
	functions map[string]funcData
}


//cтэк хранит в себе 
type memoryStack struct {
	layers []memoryLayer
	block  []int
}

func (m memoryStack) display(expression string) {
	fmt.Printf("\nVariables:\n")
	tokens := tokenize(expression[5:])
	if len(tokens) > 1 {
		m.displayVariable(tokens)
	} else {
		for lvl := len(m.layers) - 1; lvl >= m.block[len(m.block)-1]; lvl-- {
			fmt.Printf("\nVisibility level %v\n", lvl)
			for name, variable := range m.layers[lvl].variables {
				fmt.Printf("%v = %v\n", name, variable.value)
			}
		}
	}
}

func (m memoryStack) displayVariable(tokens []string) {
	fmt.Printf("\n")

	variable := ""
	for i := 0; i < len(tokens); i += 2 {
		for lvl := len(m.layers) - 1; lvl >= m.block[len(m.block)-1]; lvl-- {
			if _, ok := m.layers[lvl].variables[tokens[i]]; ok {
				variable = m.layers[lvl].variables[tokens[i]].value
				break
			}
		}
		if variable != "" {
			fmt.Printf("%v = %v\n", tokens[i], variable)
			variable = ""
		}
	}
}

func initMemory() {
	memory = memoryStack{make([]memoryLayer, 0), make([]int, 0)}
	initBlock()
}

func addLayer() {
	memory.layers = append(memory.layers, memoryLayer{make(map[string]varData), make(map[string]funcData)})
}

func removeLayer() {
	memory.layers = memory.layers[:len(memory.layers)-1]
}

func initBlock() {
	memory.block = append(memory.block, len(memory.layers))
	addLayer()
}

type funcData struct {
	arguments []string
	body      []string
}

func storeFunc(expression string) {
	tokens := tokenize(expression)
	memory.layers[len(memory.layers)-1].functions[tokens[0]] = funcData{}

	args := make([]string, 0)
	i := 2
	for ; tokens[i] != ")"; i++ {
		if tokens[i] != "," {
			args = append(args, tokens[i])
		}
	}

	body := make([]string, 0)
	i += 2
	for ; i < len(tokens)-1; i++ {
		body = append(body, tokens[i])
	}

	memory.layers[len(memory.layers)-1].functions[tokens[0]] = funcData{args, body}
}

func storeMultilineFunc(declaration []string) string {
	header := declaration[0]
	tokens := tokenize(header)

	vars := make(map[string]string, 0)
	for i := 2; i < len(tokens); i += 2 {
		vars[tokens[i]] = tokens[i]
	}

	for _, line := range declaration[1 : len(declaration)-1] {
		if !strings.Contains(line, "=") {
			return "ERROR"
		}
		variable := line[:strings.Index(line, "=")]

		tokens = tokenize(line[strings.Index(line, "=")+1 : len(line)-1])
		var builder strings.Builder
		for _, token := range tokens {
			if _, ok := vars[token]; ok {
				token = "(" + vars[token] + ")"
			}
			builder.WriteString(token)
		}

		vars[variable] = builder.String()
	}

	tokens = tokenize(declaration[len(declaration)-1])
	var body strings.Builder
	for _, token := range tokens {
		if _, ok := vars[token]; ok {
			token = "(" + vars[token] + ")"
		}
		body.WriteString(token)
	}

	storeFunc(header + ":" + body.String())

	return "ok"
}

type varData struct {
	varType string
	value   string
}

func storeVar(expression string) string {
	tokens := tokenize(expression)
	varType := ""
	switch tokens[2] {
	case "i":
		varType = "int"
	case "f":
		varType = "float"
	default:
		return "ERROR"
	}

	memory.layers[len(memory.layers)-1].variables[tokens[0]] = varData{varType, tokens[5]}

	return ""
}

func updateVariable(name string, newValue string) {
	for lvl := len(memory.layers) - 1; lvl >= memory.block[len(memory.block)-1]; lvl-- {
		variable := memory.layers[lvl].variables[name]
		variable.value = newValue
		memory.layers[lvl].variables[name] = variable
		return
	}
}

func solveInfixFunction(f funcData, tokens []string) []string {
	arguments := make([][]string, len(f.arguments))
	for i := range arguments {
		arguments[i] = make([]string, 0)
	}

	funcCounter := 0
	funcTokens := make([]string, 0)
	var innerFunc funcData
	argIndex := 0
	result := make([]string, 0)
	result = append(result, "(")

	for index, token := range tokens {
		if index == 0 || index == len(tokens)-1 {
			continue
		}
		if funcCounter == 0 {
			for lvl := len(memory.layers) - 1; lvl >= memory.block[len(memory.block)-1]; lvl-- {
				if _, ok := memory.layers[lvl].functions[token]; ok {
					innerFunc = memory.layers[lvl].functions[token]
					funcCounter++
					break
				}
			}
			if funcCounter > 0 {
				continue
			}
			if token == "," {
				argIndex++
			} else {
				arguments[argIndex] = append(arguments[argIndex], token)
			}
		} else {
			funcTokens = append(funcTokens, token)
			if token == ")" {
				funcCounter--
				if funcCounter == 1 {
					parsedTokens := solveInfixFunction(innerFunc, funcTokens)
					arguments[argIndex] = append(arguments[argIndex], parsedTokens...)

					funcTokens = make([]string, 0)
					funcCounter = 0
				}
			} else if token == "(" {
				funcCounter++
			}
		}
	}

	argsMap := make(map[string][]string)
	for i := range f.arguments {
		argsMap[f.arguments[i]] = arguments[i]
	}

	for _, token := range f.body {
		if arg, ok := argsMap[token]; ok {
			result = append(result, "(")
			result = append(result, arg...)
			result = append(result, ")")
		} else {
			result = append(result, token)
		}
	}
	result = append(result, ")")

	return result
}

func handleMinus(prev string) string {
	if prev == "(" || prev == "/" || prev == "*" || prev == "+" || prev == "-" {
		return "~"
	}
	return "-"
}

func findVariable(token string) string {
	for lvl := len(memory.layers) - 1; lvl >= memory.block[len(memory.block)-1]; lvl-- {
		if variable, ok := memory.layers[lvl].variables[token]; ok {
			return variable.value
		}
	}
	return "NO"
}

func tokenize(expression string) []string {
	tokens := make([]string, 0)
	var token strings.Builder

	for _, c := range expression {
		if c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')' ||
			c == ',' || c == ':' || c == ';' || c == '=' {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
			tokens = append(tokens, string(c))
		} else {
			token.Grow(1)
			token.WriteRune(c)
		}
	}

	for i, token := range tokens {
		if token == "-" {
			if i == 0 {
				tokens[i] = "~"
			} else {
				tokens[i] = handleMinus(tokens[i-1])
			}
		}
	}
	return tokens
}

func evaluateExpression(lineNum int, expression string, notation string) string {
	tokens := make([]string, 0)
	tokens = append(tokens, "(")
	tokens = append(tokens, tokenize(expression)...)
	tokens = append(tokens[:len(tokens)-1], ")")

	rpn := make([]string, 0)
	switch notation {
	case "prefix":

	case "infix":
		tokens = solveInfixFunction(memory.layers[0].functions["expression"], tokens)

		stack := make([]string, 0)

		for _, token := range tokens {
			if _, ok := strconv.ParseFloat(token, 64); ok == nil {
				rpn = append(rpn, token)
			} else if findVariable(token) != "NO" {
				rpn = append(rpn, findVariable(token))
			} else if token == "~" || token == "+" || token == "-" || token == "*" || token == "/" || token == "(" || token == ")" {
				switch token {
				case "~", "(":
					stack = append(stack, token)
				case "*", "/":
					for len(stack) > 0 && stack[len(stack)-1] == "~" {
						rpn = append(rpn, stack[len(stack)-1])
						stack = stack[:len(stack)-1]
					}
					stack = append(stack, token)
				case "+", "-":
					for len(stack) > 0 && (stack[len(stack)-1] == "~" || stack[len(stack)-1] == "*" || stack[len(stack)-1] == "/") {
						rpn = append(rpn, stack[len(stack)-1])
						stack = stack[:len(stack)-1]
					}
					stack = append(stack, token)
				case ")":
					for len(stack) > 0 {
						if stack[len(stack)-1] == "(" {
							stack = stack[:len(stack)-1]
							break
						}
						rpn = append(rpn, stack[len(stack)-1])
						stack = stack[:len(stack)-1]
					}
				}
			} else {
				return "[" + strconv.Itoa(lineNum) + "] ERROR: unknown token: " + token
			}
		}
	case "postfix":

	default:
		return "[" + strconv.Itoa(lineNum) + "] ERROR: unknown notation: " + notation
	}

	index := 2
	if rpn[2] == "~" {
		index--
	}

	for len(rpn) > 1 {
		switch rpn[index] {
		case "~":
			if rpn[index-1][0] == '-' {
				rpn[index-1] = rpn[index-1][1:]
			} else {
				rpn[index-1] = "-" + rpn[index-1]
			}
			rpn = append(rpn[:index], rpn[index+1:]...)
			index--
		case "+", "-", "*", "/":
			a, ok := strconv.ParseFloat(rpn[index-2], 64)
			if ok != nil {
				return "[" + strconv.Itoa(lineNum) + "] ERROR: can't parse float " + rpn[index-2]
			}
			b, ok := strconv.ParseFloat(rpn[index-1], 64)
			if ok != nil {
				return "[" + strconv.Itoa(lineNum) + "] ERROR: can't parse float " + rpn[index-1]
			}

			switch rpn[index] {
			case "+":
				rpn[index-2] = fmt.Sprintf("%v", a+b)
			case "-":
				rpn[index-2] = fmt.Sprintf("%v", a-b)
			case "*":
				rpn[index-2] = fmt.Sprintf("%v", a*b)
			case "/":
				rpn[index-2] = fmt.Sprintf("%v", a/b)
			}
			rpn = append(rpn[:index-1], rpn[index+1:]...)
			index -= 2
		default:
			index++
		}
	}
	return strings.Join(rpn, "")
}

func RunInterpreter(filePath string) {
	initMemory()
	memory.layers[0].functions["expression"] = funcData{[]string{"x"}, []string{"x"}}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Couldn't find the file \"%v\": %v\n", filePath, err)
	} else {
		lineNum := 0
		scanner := bufio.NewScanner(file)
		inFunc := make([]string, 0)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.ReplaceAll(line, " ", "")
			lineNum++

			fmt.Printf("line %v: %v\n\tinFunc: %v\n", lineNum, line, inFunc)

			if len(inFunc) > 0 {
				if strings.Contains(line, "return") {
					inFunc = append(inFunc, line[6:])
					if storeMultilineFunc(inFunc) == "ERROR" {
						fmt.Printf("ERROR")
						return
					}
					inFunc = make([]string, 0)
				} else {
					inFunc = append(inFunc, line)
				}
			} else if line == "{" {
				addLayer()
			} else if line == "}" {
				removeLayer()
			} else if strings.Contains(line, "(i)=") || strings.Contains(line, "(f)=") {
				if storeVar(line) == "ERROR" {
					fmt.Printf("[%v] Couldn't save variable, invalid type\n", lineNum)
				}
			} else if strings.Contains(line, "print") {
				memory.display(line)
			} else if strings.Contains(line, ":") && strings.Index(line, ":") == len(line)-1 {
				inFunc = append(inFunc, line[:strings.Index(line, ":")])
			} else if strings.Contains(line, "return") {
				inFunc = append(inFunc, line)
				if storeMultilineFunc(inFunc) == "ERROR" {
					fmt.Printf("ERROR")
				}
				inFunc = make([]string, 0)
			} else if strings.Contains(line, ":") {
				storeFunc(line)
			} else if strings.Contains(line, "=") {
				var saveLine strings.Builder
				saveLine.WriteString(line[:strings.Index(line, "=")+1])
				assignation := evaluateExpression(lineNum, line[strings.Index(line, "=")+1:], "infix")
				if assignation == "ERROR" {
					fmt.Printf("[line %v] Couldn't assign value\n", lineNum)
				} else {
					saveLine.WriteString(assignation)
					updateVariable(line[:strings.Index(line, "=")], assignation)
				}
			}
		}
	}
}

func main() {
	RunInterpreter("/home/sergey/micro/3.3/3.3.txt")
}
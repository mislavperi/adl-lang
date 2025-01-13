package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/mislavperi/gem-lang/compiler"
	"github.com/mislavperi/gem-lang/lexer"
	"github.com/mislavperi/gem-lang/object"
	"github.com/mislavperi/gem-lang/parser"
	"github.com/mislavperi/gem-lang/vm"
)

const PROMPT = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	for index, builtin := range object.Builtins {
		symbolTable.DefineBuiltin(index, builtin.Name)
	}

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		compiler := compiler.NewWithState(symbolTable, constants)
		err := compiler.Compile(program)
		if err != nil {
			fmt.Fprint(out, "Complidation failed:\n", err)
			continue
		}

		code := compiler.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalStore(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprint(out, "Execution of bytecode failed:\n", err)
			continue
		}

		stackTop := machine.LastPoppedStackElem()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")

	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Woops! We ran into some trouble here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

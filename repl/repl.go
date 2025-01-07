package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/mislavperi/gem-lang/compiler"
	"github.com/mislavperi/gem-lang/lexer"
	"github.com/mislavperi/gem-lang/parser"
	"github.com/mislavperi/gem-lang/vm"
)

const PROMPT = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

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

		compiler := compiler.New()
		err := compiler.Compile(program)
		if err != nil {
			fmt.Fprint(out, "Complidation failed:\n", err)
			continue
		}

		machine := vm.New(compiler.Bytecode())
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
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

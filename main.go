package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mislavperi/adl-lang/compiler"
	"github.com/mislavperi/adl-lang/lexer"
	"github.com/mislavperi/adl-lang/parser"
	"github.com/mislavperi/adl-lang/repl"
	"github.com/mislavperi/adl-lang/representation"
	symboltable "github.com/mislavperi/adl-lang/symbol_table"
	"github.com/mislavperi/adl-lang/vm"
)

func main() {
	fmt.Print("Hello! This is the ADl programming language!\n")

	if len(os.Args) > 1 {
		file := os.Args[1]
		if err := executeFile(file); err != nil {
			fmt.Printf("Error executing file: %s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Feel free to type in some commands\n")
		repl.Start(os.Stdin, os.Stdout)
	}
}

func executeFile(filename string) error {
	if filepath.Ext(filename) != ".adl" {
		return fmt.Errorf("invalid file extension, expected .adl")
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	line := string(content)
	l := lexer.New(line)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(os.Stdout, p.Errors())
		return fmt.Errorf("parsing errors")
	}

	constants := []representation.Representation{}
	globals := make([]representation.Representation, vm.GlobalsSize)
	symbolTable := symboltable.NewSymbolTable()

	for index, builtin := range representation.Builtins {
		symbolTable.DefineBuiltin(index, builtin.Name)
	}

	compiler := compiler.NewWithState(symbolTable, constants)
	err = compiler.Compile(program)
	if err != nil {
		return fmt.Errorf("compilation failed: %v", err)
	}

	code := compiler.Bytecode()
	constants = code.Constants

	machine := vm.NewWithGlobalStore(code, globals)
	err = machine.Run()
	if err != nil {
		return fmt.Errorf("execution of bytecode failed: %v", err)
	}

	stackTop := machine.LastPoppedStackElem()
	fmt.Println(stackTop.Inspect())

	return nil
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintln(out, "Woops! We ran into some trouble here!")
	fmt.Fprintln(out, " parser errors:")
	for _, msg := range errors {
		fmt.Fprintf(out, "\t%s\n", msg)
	}
}

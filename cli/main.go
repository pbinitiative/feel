package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pbinitiative/feel"
	"io"
	"os"
)

func main() {
	cliFlags := flag.NewFlagSet("feel", flag.ExitOnError)

	pCmdStr := cliFlags.String("c", "", "feel script as string")
	pDumpAST := cliFlags.Bool("ast", false, "dump ast tree only")

	cliFlags.Parse(os.Args[1:])

	input := *pCmdStr
	if input == "" {

		if cliFlags.NArg() <= 0 {
			// fmt.Fprintln(os.Stderr, "no input file")
			// os.Exit(1)
			// read from stdin
			reader := bufio.NewReader(os.Stdin)
			data, err := io.ReadAll(reader)
			if err != nil {
				panic(err)
			}
			input = string(data)
		} else {
			data, err := os.ReadFile(cliFlags.Args()[0])
			if err != nil {
				panic(err)
			}
			input = string(data)
		}
	}
	if *pDumpAST {
		ast, err := feel.ParseString(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error, %s\n", err)
			os.Exit(1)
			// panic(err)
		}
		fmt.Println(ast.Repr())
	} else {
		res, err := feel.EvalString(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "eval error, %s\n", err)
			os.Exit(1)
		}
		bytes, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			//panic(err)
			fmt.Fprintf(os.Stderr, "dump error, %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(bytes))
	}
}

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/arregist97/Hydro-Compiler/tokenizer"
	"github.com/arregist97/Hydro-Compiler/parser"
	"github.com/arregist97/Hydro-Compiler/generator"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Incorrect Usage. Expected:")
		fmt.Println("main.go <filename>")
		return
	}

        fileName := os.Args[1]
	
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var empty []string
	tokens := tokenizer.RecTokenize(string(content), empty)
	fmt.Println(tokens)
	store := parser.NewNodeStore()
	tree := parser.BuildTokenTree(store, tokens, false)
	fmt.Println("\nToken Tree:")
	tree.PrintTokenTree()

	buffer, err := generator.Generate(tree)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buffer)

	fileName = filepath.Base(fileName)
	re := regexp.MustCompile(`\.[^.]+$`)
	baseName := re.ReplaceAllString(fileName, "")
	directory := "../build/"
	newFileName := baseName + ".asm"
	buildPath := directory + newFileName

	newFile, err := os.Create(buildPath)
	if err != nil {
		fmt.Println("failed to create new file: ", err)
	}
	defer newFile.Close()

	_, err = newFile.WriteString(buffer)
	if err != nil {
		fmt.Println("failed to write to new file: ", err)
	}

	oFileName := baseName + ".o"
	fmt.Println("nasm -felf64", newFileName)
	nasmCmd := exec.Command("nasm", "-felf64", newFileName)

	nasmCmd.Dir = "../build"

	nasmCmd.Stdout = os.Stdout
	nasmCmd.Stderr = os.Stderr

	// Run the nasm command
	err = nasmCmd.Run()
	if err != nil {
		log.Fatalf("nasm command execution failed: %v", err)
	}

	// Step 2: Run ld command
	fmt.Println("Running ld test.o -o test")

	// Create the ld command
	ldCmd := exec.Command("ld", oFileName, "-o", baseName)
	ldCmd.Dir = "../build"
	ldCmd.Stdout = os.Stdout
	ldCmd.Stderr = os.Stderr

	// Run the ld command
	err = ldCmd.Run()
	if err != nil {
		log.Fatalf("ld command execution failed: %v", err)
	}

	fmt.Println("Successfully assembled and linked the program.")

}

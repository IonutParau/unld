package main

import (
    "os"
    "fmt"
    "github.com/IonutParau/unld/disassemble"
)

func main() {
    if len(os.Args) == 1 {
        fmt.Printf("Usage: %s [options]\n", os.Args[0])
        options := []string{
            "--empty - Empties the current object context",
            "--section [section] - Switches the current section to section. By default, the current section is .text",
            "-s - Alias for --section",
            "--add [symbol] - Adds [symbol] from the current section to the current object context",
            "-a - Alias for --add",
            "--remove [symbol] - Removes [symbol] from the current section fom the current object context",
            "-r - Alias for --remove",
            "--global [global] - Makes the object file define the global instead of defining it as an external symbol",
            "-g - Alias for --global",
            "--output [file] - Outputs an object file generated from the current object context and puts it in [file].",
            "\tThis also resets the current object context to contain all sections and symbols from the executable (except the insignificant ones)",
            "-o - Alias for --output",
        }
        for _, option := range options {
            fmt.Printf("\t%s\n", option)
        }
        os.Exit(1)
    }

    input := os.Args[1]
    files, err := disassemble.GetLinkedFiles(input)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    sections, err := disassemble.GetDumpedAssembly(input)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    symbols := disassemble.FindExternSymbols(sections)
    sections = disassemble.RemoveJunk(sections)

    rodata, err := disassemble.GetDumpedReadonlyData(input)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    
    globaldata, err := disassemble.GetDumpedGlobalData(input)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    
    sections = disassemble.MonkeyPatchAssembly(sections, globaldata, rodata, symbols)

    objectContext := disassemble.Object{
        Globals: globaldata,
        Literals: rodata,
        Sections: sections,
    }
    baseContext := objectContext
    currentSection := ".text"

    for i := 2; i < len(os.Args); i++ {
        arg := os.Args[i]

        if arg == "--empty" {
            objectContext = objectContext.Empty()
            continue
        }

        if arg == "--section" || arg == "-s" {
            currentSection = os.Args[i+1]
            i++
            continue
        }

        if arg == "--add" || arg == "-a" {
            symbol := os.Args[i+1]
            i++
            objectContext = objectContext.TakeSymbolFrom(symbol, currentSection, baseContext)
            continue
        }
        
        if arg == "--remove" || arg == "-r" {
            symbol := os.Args[i+1]
            i++
            objectContext = objectContext.RemoveSymbol(symbol, currentSection)
            continue
        }

        if arg == "--global" || arg == "-g" {
            symbol := os.Args[i+1]
            i++
            objectContext = objectContext.IncludeGlobal(symbol)
        }
        
        if arg == "--output" || arg == "-o" {
            file := os.Args[i+1]
            i++
            // Output needs the linked files (obviously)
            err := objectContext.Output(file, files, objectContext.AddNecessarySymbols(baseContext, symbols))
            if err != nil {
                fmt.Println(err)
                os.Exit(1)
            }
            objectContext = baseContext
            continue
        }
    }
}

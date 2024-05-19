package disassemble

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Object struct {
    Globals []Data
    Literals []Data
    Sections []Section
}

func (o Object) Empty() Object {
    return Object {
        o.Globals,
        o.Literals,
        make([]Section, 0, len(o.Sections)),
    }
}

func (o Object) RemoveSection(name string) Object {
    sections := make([]Section, 0, len(o.Sections))

    for _, section := range o.Sections {
        if section.Name == name {
            continue
        }

        sections = append(sections, section)
    }

    return Object {
        o.Globals,
        o.Literals,
        sections,
    }
}

func (o Object) RemoveSymbol(name string, section string) Object {
    sections := make([]Section, 0, len(o.Sections))

    for _, sec := range o.Sections {
        if sec.Name == section {
            funcs := make([]AssemblyFunction, 0, len(sec.Funcs))
            for _, fun := range sec.Funcs {
                if fun.Name == name {
                    continue
                }

                funcs = append(funcs, fun)
            }
            sections = append(sections, Section{sec.Name, funcs})
            continue
        }

        sections = append(sections, sec)
    }

    return Object {
        o.Globals,
        o.Literals,
        sections,
    }
}

func (o Object) HasSection(name string) bool {
    for _, section := range o.Sections {
        if section.Name == name {
            return true
        }
    }

    return false
}

func (o Object) HasSymbol(name string, section string) bool {
    for _, sec := range o.Sections {
        if sec.Name == name {
            for _, fun := range sec.Funcs {
                if fun.Name == name {
                    return true
                }
            }
        }
    }

    return false
}

func (o Object) TakeSectionFrom(name string, source Object) Object {
    if o.HasSection(name) {
        return o
    }

    sections := make([]Section, 0, len(o.Sections)+1)
    for _, sec := range o.Sections {
        sections = append(sections, sec)
    }
    for _, sec := range source.Sections {
        if sec.Name == name {
            sections = append(sections, sec)
        }
    }

    return Object {
        o.Globals,
        o.Literals,
        sections,
    }
}

func (o Object) TakeSymbolFrom(name string, section string, source Object) Object {
    if o.HasSymbol(name, section) {
        return o
    }

    if !o.HasSection(section) {
        o.Sections = append(o.Sections, Section{section, []AssemblyFunction{}})
    }

    sections := make([]Section, 0, len(o.Sections))

    for _, sec := range o.Sections {
        if sec.Name == section {
            found := false
            for _, sourceSec := range source.Sections {
                if sourceSec.Name == section {
                    for _, fun := range sourceSec.Funcs {
                        if fun.Name == name {
                            funcs := make([]AssemblyFunction, 0, len(sec.Funcs))

                            for _, oFun := range sec.Funcs {
                                funcs = append(funcs, oFun)
                            }

                            funcs = append(funcs, fun)
                            sections = append(sections, Section{section, funcs})
                            found = true
                            break
                        }
                    }
                }
            }

            if found {
                continue
            }
        }

        sections = append(sections, sec)
    }

    return Object {
        o.Globals,
        o.Literals,
        sections,
    }
}

func (o Object) IncludeGlobal(name string) Object {
    globals := make([]Data, 0, len(o.Globals))

    for _, global := range o.Globals {
        if global.Name == name {
            global.Extern = false // this will instead use the binary info
        }
        globals = append(globals, global)
    }

    return Object{
        globals,
        o.Literals,
        o.Sections,
    }
}

func (o Object) Trim() Object {
    globals := make([]Data, 0, len(o.Globals))
    literals := make([]Data, 0, len(o.Literals))

    globalLoop: for _, global := range o.Globals {
        for _, sec := range o.Sections {
            for _, fun := range sec.Funcs {
                for _, line := range fun.Content {
                    if strings.Contains(line, "[rel " + global.Name + "]") || strings.Contains(line, "[rel " + global.Name + "+") {
                        globals = append(globals, global)
                        continue globalLoop
                    }
                }
            }
        }
    }
    literalLoop: for _, literal := range o.Literals {
        for _, sec := range o.Sections {
            for _, fun := range sec.Funcs {
                for _, line := range fun.Content {
                    if strings.Contains(line, literal.Name) {
                        literals = append(literals, literal)
                        continue literalLoop
                    }
                }
            }
        }
    }

    return Object {
        globals,
        literals,
        o.Sections,
    }
}

func (o Object) TrimSymbols(symbols []string) []string {
    used := make([]string, 0, len(symbols))

    symbolLoop: for _, symbol := range symbols {
        for _, sec := range o.Sections {
            for _, fun := range sec.Funcs {
                for _, line := range fun.Content {
                    if strings.Contains(line, symbol) {
                        used = append(used, symbol)
                        continue symbolLoop 
                    }
                }
            }
        }
    }

    return used
}

func (o Object) AddUnusedSymbols(symbols []string) []string {
    used := make([]string, 0, len(symbols))

    for _, sec := range o.Sections {
        for _, fun := range sec.Funcs {
            for _, line := range fun.Content {
                for _, symbol := range symbols {
                    if strings.Contains(line, symbol) {
                        found := false
                        for _, u := range used {
                            if u == symbol {
                                found = true
                                break
                            }
                        }
                        if !found {
                            used = append(used, symbol)
                        }
                    }
                }
            }
        }
    }

    return used
}

func (o Object) AddNecessarySymbols(in Object, symbols []string) []string {
    used := make([]string, 0, len(symbols))

    for _, symbol := range symbols {
        used = append(used, symbol)
    }

    for _, inSec := range in.Sections {
        for _, inFun := range inSec.Funcs {
            for _, oSec := range o.Sections {
                for _, oFun := range oSec.Funcs {
                    if oFun.Name == inFun.Name {
                        continue
                    }
                    for _, oLine := range oFun.Content {
                        if strings.Contains(oLine, inFun.Name) {
                            found := false
                            for _, u := range used {
                                if u == inFun.Name {
                                    found = true
                                }
                            }
                            if !found {
                                used = append(used, inFun.Name)
                            }
                        }
                    }
                }
            }
        }
    }

    return used
}

func (o Object) Output(filepath string, files []string, symbols []string) error {
    file, err := os.CreateTemp("", "unld_asm_")
    //file, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer os.Remove(file.Name()) 
    //defer file.Close()

    o = o.Trim()
    symbols = o.TrimSymbols(symbols)
    symbols = o.AddUnusedSymbols(symbols)

    for _, symbol := range symbols {
        fmt.Fprintln(file, "extern", symbol)
    }
    if len(o.Globals) > 0 {
        fmt.Fprintln(file, "section .bss")
        for _, global := range o.Globals {
            if global.Extern {
                fmt.Fprintf(file, "extern %s\n", global.Name)
            } else {
                fmt.Fprintf(file, "global %s\n%s:\n", global.Name, global.Name)
                fmt.Fprintf(file, "\tresb %d\n", len(global.Data))
            }
        }
    }
    if len(o.Literals) > 0 {
        fmt.Fprintln(file, "section .rodata")
        for _, literal := range o.Literals {
            fmt.Fprintf(file, "%s:\n", literal.Name)
            fmt.Fprintf(file, "\tdb ")
            for i := 0; i < len(literal.Data); i++ {
                fmt.Fprintf(file, "%d", literal.Data[i])
                if i == len(literal.Data)-1 {
                    fmt.Fprint(file, "\n")
                } else {
                    fmt.Fprintf(file, ",")
                }
            }
        }
    }
    for _, section := range o.Sections {
        fmt.Fprintf(file, "section %s\n", section.Name)
        for _, fun := range section.Funcs {
            fmt.Fprintf(file, "global %s\n%s:\n", fun.Name, fun.Name)
            for _, line := range fun.Content {
                fmt.Fprintf(file, "\t%s\n", line)
            }
        }
    }

    assembling := exec.Command("nasm", "-felf64", file.Name(), "-o", filepath)
    out, err := assembling.CombinedOutput()

    if err != nil {
        return errors.New(string(out))
    }

    return nil
}

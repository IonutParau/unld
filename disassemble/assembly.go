package disassemble

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type AssemblyFunction struct {
    Name string
    Content []string
}

type Section struct {
    Name string
    Funcs []AssemblyFunction
}

type Data struct {
    Name string
    Location int
    Data []byte
    Extern bool
}

func GetDumpedData(file string, segment string) ([]Data, error) {
    cmd := exec.Command("objdump", "-d", file, "-j", segment, "-z")

    buf, err := cmd.CombinedOutput()

    if err != nil {
        return nil, err
    }

    content := string(buf)

    lines := strings.Split(strings.TrimSpace(content), "\n")
    for i := range lines {
        lines[i] = strings.TrimSpace(lines[i])
    }

    data := []Data{}

    for _, line := range lines {
        if strings.Contains(line, "file format") {
            continue
        }

        if strings.HasPrefix(line, "Disassembly of section ") {
            continue
        }

        if name, ok := strings.CutSuffix(line, ":"); ok {
            loc, err := strconv.ParseUint(name[0:strings.Index(name, " ")], 16, 64)
            if err != nil {
                return nil, err
            }
            name = name[strings.Index(name, "<")+1:len(name)-1]
            data = append(data, Data{name, int(loc), []byte{}, false})
        } else if strings.ContainsRune(line, rune(9)) {
            rawBytes := strings.Split(strings.Split(line, string(rune(9)))[1], " ")
            bytes := []byte{}
            for _, b := range rawBytes {
                if len(b) == 0 {
                    break
                }
                parsed, err := strconv.ParseUint(b, 16, 8)
                if err != nil {
                    return nil, err
                }
                bytes = append(bytes, byte(parsed))
            }
            last := len(data)-1
            data[last].Data = append(data[last].Data, bytes...)
        }
    }

    return data, nil
}

func GetDumpedReadonlyData(file string) ([]Data, error) {
    return GetDumpedData(file, ".rodata")
}

func GetDumpedGlobalData(file string) ([]Data, error) {
    // All globals are extern by default because I said so.
    global, err := GetDumpedData(file, ".bss")
    if err != nil {
        return nil, err
    }

    for i, g := range global {
        g.Extern = true
        global[i] = g
    }

    return global, nil
}

func FindExternSymbols(sections []Section) []string {
    symbols := []string{}

    for _, section := range sections {
        if section.Name == ".plt" {
            for _, fun := range section.Funcs {
                if strings.Contains(fun.Name, "-") {
                    continue // assume it is insignificant thunk stuff
                }

                if name, ok := strings.CutSuffix(fun.Name, "@plt"); ok {
                    symbols = append(symbols, name)
                }
            }
        }
    }

    return symbols
}

func RemoveJunk(sections []Section) []Section {
    useful := make([]Section, 0, len(sections))

    for _, section := range sections {
        if section.Name == ".init" || section.Name == ".plt" || section.Name == ".fini" {
            continue
        }

        funcs := make([]AssemblyFunction, 0, len(section.Funcs))

        for _, fun := range section.Funcs {
            // libc will provide it anyways, so who cares?
            if fun.Name == "_start" && section.Name == ".text" {
                continue
            }

            funcs = append(funcs, fun)
        }

        useful = append(useful, Section{section.Name, funcs})
    }

    return useful
}

func GetDumpedAssembly(file string) ([]Section, error) {
    cmd := exec.Command("objdump", "-d", "-M", "intel noprefix", "--no-show-raw-insn", file)

    buf, err := cmd.CombinedOutput()

    if err != nil {
        return nil, err
    }

    content := string(buf)

    lines := strings.Split(strings.TrimSpace(content), "\n")
    for i := range lines {
        lines[i] = strings.TrimSpace(lines[i])
    }
    
    sections := []Section{}

    for _, line := range lines {
        if len(line) == 0 {
            continue
        }

        if strings.Contains(line, "file format") {
            continue
        }

        if section, ok := strings.CutPrefix(line, "Disassembly of section "); ok {
            section := section[:len(section)-1]

            sections = append(sections, Section{section, []AssemblyFunction{}});
        } else if funcName, ok := strings.CutSuffix(line, ":"); ok {
            funcName = funcName[strings.Index(funcName, "<")+1:len(funcName)-1]
            last := len(sections)-1
            sections[last].Funcs = append(sections[last].Funcs, AssemblyFunction{funcName, []string{}})
        } else if strings.ContainsRune(line, rune(9)) {
            code := line[strings.IndexRune(line, rune(9))+1:]
            code = strings.ReplaceAll(code, " PTR", "")
            last := len(sections)-1
            lastFun := len(sections[last].Funcs)-1
            sections[last].Funcs[lastFun].Content = append(sections[last].Funcs[lastFun].Content, code)
        }
    }

    return sections, nil
}

func MonkeyPatchAssembly(sections []Section, globals []Data, literals []Data, symbols []string) []Section {
    output := make([]Section, 0, len(sections))

    for _, section := range sections {
        funcs := make([]AssemblyFunction, 0, len(section.Funcs))

        for _, fun := range section.Funcs {
            code := make([]string, 0, len(fun.Content))

            for _, line := range fun.Content {
                if !strings.Contains(line, "# ") {
                    if strings.Contains(line, "<") {
                        symbol := line[strings.Index(line, "<")+1:strings.Index(line, ">")]
                        if ext, ok := strings.CutSuffix(symbol, "@plt"); ok {
                            symbol = ext
                        }
                        cmd := line[0:strings.Index(line, " ")]
                        code = append(code, fmt.Sprintf("%s %s", cmd, symbol))
                        continue
                    }
                    code = append(code, line)
                    continue
                }

                assembly := line[0:strings.Index(line, "# ")]
                comment := line[strings.Index(line, "# ")+2:]

                symbol := comment[strings.Index(comment, "<")+1:strings.Index(comment, ">")]
                // Relative address, assembly needs patching
                if strings.Contains(assembly, "[rip") {
                    beforeRip := assembly[0:strings.Index(assembly, "[rip")]
                    afterRip := assembly[strings.Index(assembly, "]")+1:]

                    assembly = fmt.Sprintf("%s[rel %s]%s", beforeRip, symbol, afterRip)
                }

                code = append(code, assembly)
            }

            funcs = append(funcs, AssemblyFunction{fun.Name, code})
        }

        output = append(output, Section{section.Name, funcs})
    }

    return output
}

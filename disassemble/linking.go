package disassemble

import (
    "os/exec"
    "strings"
    "errors"
)

func GetLinkedFiles(file string) ([]string, error) {
    dynlink, err := IsDynamicallyLinked(file)
    if err != nil {
        return nil, err
    }

    if !dynlink {
        return []string{}, nil
    }

    cmd := exec.Command("ldd", file)
    buf, err := cmd.CombinedOutput()

    if err != nil {
        return nil, errors.New(string(buf)) 
    }

    content := string(buf)

    files := strings.Split(strings.TrimSpace(content), "\n")
    interpreter, err := GetInterpreter(file)
    if err != nil {
        return nil, err
    }

    for i := 0; i < len(files); i++ {
        files[i] = strings.TrimSpace(files[i])
    }

    withoutInterpreter := make([]string, len(files)-1)
    j := 0
    for i := 0; i < len(files); i++ {
        // because of =>
        if strings.HasPrefix(files[i], interpreter) {
            continue
        }
        withoutInterpreter[j] = files[i][0:strings.Index(files[i], " (")]
        j++
    }

    return withoutInterpreter, nil
}

func IsDynamicallyLinked(file string) (bool, error) {
    cmd := exec.Command("file", file)

    buf, err := cmd.CombinedOutput()

    if err != nil {
        return false, errors.New(string(buf))
    }

    out := string(buf)

    return strings.Contains(out, "dynamically linked"), nil
}

func GetInterpreter(file string) (string, error) {
    cmd := exec.Command("file", file)

    buf, err := cmd.CombinedOutput()

    if err != nil {
        return "", errors.New(string(buf))
    }

    out := strings.Split(string(buf), ", ")

    for _, part := range out {
        if interpreter, ok := strings.CutPrefix(part, "interpreter "); ok {
            return interpreter, nil
        }
    }

    return "", errors.New("Unable to find interpreter")
}

func IsStaticallyLinked(file string) (bool, error) {
    x, err := IsDynamicallyLinked(file)

    return !x, err
}

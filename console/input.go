package console

import (
  "fmt"
  "strings"
)

type PrintFunction func(attempts int) error

func ExpectIntRange(min int, max int, print PrintFunction) (value int, err error) {
  attempts := 0
  for ; ; {
    if err = print(attempts); err != nil {
      return
    }

    _,err = fmt.Scanf("%d\n", &value)

    if err == nil && value >= min && value <= max {
      return
    }
    attempts++
  }

  return
}

func ExpectAnyString(print PrintFunction) (value string, err error) {
  attempts := 0
  for ; ; {
    if err = print(attempts); err != nil {
      return
    }
    _,err = fmt.Scanln(&value)

    if err == nil {
      return
    }
    attempts++
  }

  return
}

func ExpectRestrictedString(values []string, print PrintFunction) (value string, err error) {
  attempts := 0
  valid := false
  for ; ; {
    if err = print(attempts); err != nil {
      return
    }

    fmt.Printf(": ")
    _,err = fmt.Scanln(&value)

    value = strings.TrimSpace(value)
    for _,v :=range values {
      if v == value {
        valid = true
        break
      }
    }

    if valid {
      return
    }

    attempts++
  }

  return
}

func Prompt(prompt string) PrintFunction {
  return func (attempts int) error {
    fmt.Printf(prompt)
    return nil
  }
}

func ListPrompt(title string, items ...string) PrintFunction {
  fmt.Println(len(items))
  return func (attempts int) error {
    fmt.Println(title)
    for i := 0; i < len(items); i++ {
      fmt.Printf("(%03d) %s\n", i+1, items[i])
    }
    fmt.Printf("[%d-%d]: ", 1, len(items))
    return nil
  }
}

func KeyValuePrompt(title string, keys []string, values []string) PrintFunction {
  return func (attempts int) error {
    fmt.Println(title)
    for i := 0; i < len(keys); i++  {
      fmt.Printf("(%s) %s\n", keys[i], values[i])
    }
    return nil
  }
}

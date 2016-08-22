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

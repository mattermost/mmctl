package prompt

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/isacikgoz/prompt/term"
)

type Selection struct {
	prompt *Prompt
	result string
}

func NewSelection(question string, answers []string, footnote string, maxLineSize int) (*Selection, error) {
	list, err := NewList(answers, maxLineSize)
	if err != nil {
		return nil, err
	}

	sel := &Selection{}

	selFn := func(item interface{}) error {
		sel.result = fmt.Sprintf("%s", item)
		sel.prompt.Stop()
		return nil
	}

	infoFn := func(item interface{}) [][]term.Cell {
		i := term.Cprint(footnote, color.FgRed)
		return [][]term.Cell{i}
	}

	sel.prompt = Create(question, &Options{LineSize: maxLineSize}, list,
		WithSelectionHandler(selFn), WithInformation(infoFn))

	return sel, nil
}

func (s *Selection) Run() (string, error) {
	err := s.prompt.Run(context.Background())
	return s.result, err
}

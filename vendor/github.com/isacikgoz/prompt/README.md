# Prompt

Prompt for cli apps. Currently in development, use with care.

## Use

```Go
sel, err := prompt.NewSelection("are you sure?", []string{"yes", "no"}, "* pick wisely", 2)
if err != nil {
	log.Fatal(err)
}

answer, err := sel.Run()
if err != nil {
	log.Fatal(err)
}

log.Printf("selected %q\n", answer)
```
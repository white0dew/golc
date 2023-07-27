---
title: Prompt Template
description: All about prompt templates.
weight: 10
---
```go
import "github.com/hupe1980/golc/prompt"

template := `You are a naming consultant for new companies.
What is a good name for a company that makes {{.product}}?`

pt = promt.NewTemplate(template)

p, err := pt.Format(map]string]any{
    "product": "colorful socks",
})
if err != nil {
   // Error handling
}

fmt.Println(p)
```
Output:
```text
You are a naming consultant for new companies.
What is a good name for a company that makes colorful socks?
```
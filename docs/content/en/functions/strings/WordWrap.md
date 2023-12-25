---
title: strings.WordWrap
description: Returns the given string, split by newlines such that each line does not exceed the given length.
categories: []
keywords: []
action:
  aliases: []
  related: []
  returnType: string
  signatures: [strings.WordWrap LENGTH INPUT]
---

```go-html-template
{{ strings.WordWrap 15 "Able was I, ere I saw Elba." }} → Able was I, ere\nI saw Elba.
{{ strings.WordWrap 80 "Able was I, ere\nI saw Elba." }} → Able was I, ere I saw Elba.
```

By example, to wrap a page's [`RawContent`], split it into blocks, then wrap each block:

```go-html-template
{{ range split (replace .RawContent "\r" "\n") "\n\n" }}
  {{- printf "%s\n" (strings.WordWrap 80 .) }}
{{ end }}
```

To wrap the raw content with shortcodes evaluated, use the [`RenderShortcodes`] method instead:

```go-html-template
{{ range split (replace .RenderShortcodes "\r" "\n") "\n\n" }}
  {{- printf "%s\n" (strings.WordWrap 80 .) }}
{{ end }}
```

[`RawContent`]: /methods/page/rawcontent
[`RenderShortcodes`]: /methods/page/rendershortcodes

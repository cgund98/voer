package frontend

// Doing this to avoid having to ship a static file with the CLI
var css string = `

@plugin "daisyui" {
  themes: light --default, dark --prefersdark, cupcake;
}

[data-theme="dark"] {
  * {
	font-family: "Inter", sans-serif;
  }
}

dialog {
  background-color: transparent;
}

[x-cloak] { display: none !important; }

pre code {
  background-color: transparent !important;
}
`

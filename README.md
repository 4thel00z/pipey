# pipey 🚰

![Pipey Logo](./pipey.jpeg)

`pipey` (pronounced /paɪpi/ like "pipe-ee") is a powerful utility that bridges the world of UNIX pipes and HTTP servers. With `pipey`, you can expose any data stream from a named pipe over HTTP, allowing for a myriad of applications, especially in creating non-blocking UIs with bash using Go binaries.

## Motivation 🌟

While UNIX pipes are incredibly powerful, they're limited to inter-process communication on the same machine. What if you wanted to expose the data from a pipe to other systems or services over HTTP? That's where `pipey` comes into play.

Moreover, in the era of interactive CLIs and TUIs, there's a growing need to integrate traditional shell scripting with modern UI paradigms. Tools like `charty` allow us to create visually appealing interfaces right from the terminal. By using `pipey`, we can further enhance these UIs by fetching data asynchronously over HTTP, ensuring our UI remains responsive and snappy.

Imagine crafting non-blocking UIs using bash and Go binaries, where the data is sourced from various scripts or processes via `pipey`. The possibilities are endless!

## Installation 🛠

To install `pipey`, use the following `go install` command:

```bash
go install github.com/4thel00z/pipey/...@latest
```

Ensure your Go bin directory (usually $HOME/go/bin) is in your PATH to access the pipey command.


## Usage 🚀

```bash
pipey [PIPE_NAME] --host [HOST] --port [PORT] --timeout [SECONDS]
```

For more details and options, refer to the command help:

```bash
pipey --help
```

## Contributing 🤝

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

## License 📄

This project is licensed under the GPL-3 license.

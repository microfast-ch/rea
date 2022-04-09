# rea
Rea is a document renderer that makes your document generation easy.
It can process [ODF](https://www.libreoffice.org/discover/what-is-opendocument/)
and [OOXML](https://support.microsoft.com/en-gb/office/open-xml-formats-and-file-name-extensions-5200d93c-3449-4380-8e11-31ef14555b18) files
and uses [Lua](https://www.lua.org/) as the templating language.

## Usage
Rea has two main functions:

- Templating: Filling out a document or template with data
- Rendering: Converting a document to a archivable format (e.g. PDF)

Usually you will have a template document that contains instructions on how data
is connected to the document structure. During the `template`-step you will bring
these two parts together to have a filled out document. This document you can
further edit or use the `render`-step to create a PDF from it.

### Templating
Rea uses Lua as it's templating language. This allows you to have a well known,
simple and convenient yet powerful processing engine.

It works by having your (template) document as is but introducing two text blocks
that have a special function:

- `[[ foo ]]`: This is a code block, everything between `[[` and `]]` is interpreted as lua code
- `[# bar #]`: This is a print block, everything between `[#` and `#]` is printed out into the document. It's a shorthand for calling `Print(bar)` in a code block.

Let's take this example document:
TODO: Picture of a document

Here you see that we are using ... TODO: Explaination and picture of result

#### Creating a template
TODO:
- Lua introduction and examples

#### Passing data to the document
You can pass data to the template by having an input file as yaml. It should contain
two top level keys `data` and `metadata`, where you are free to define your data structure.
The `metadata` key is special as it will be used to set the documents metadata like author.

Example:
```yaml
metadata:
  author: "John Doe"
data:
  customer:
    firstname: "Sue"
    lastname: "Chang"
  items:
  - Apple
  - Banana
  - Lemon
  greeting: "Hello Sue!"
```

The fields from `data` can be accessed directly in your document (e.g. `[# customer.firstname #]`)
whereas `metadata` values needs to be accessed through the `metadata`-prefix (e.g. `[# metadata.author #]`).

#### Generate templated document
```plaintext
Process a template document to generate a filled out document

Usage:
  rea template [flags]

Flags:
  -b, --bundle string     tar file to which the job bundle should be written
  -d, --debug             write debug information to job bundle
  -h, --help              help for template
  -i, --input string      data file (default "data.yaml")
  -o, --output string     output document (default "document.odt")
  -t, --template string   template document (default "template.ott"
```

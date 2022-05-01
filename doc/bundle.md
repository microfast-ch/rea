# Bundle
rea creates a bundle during the processing. It can be used for further processing or to debug issues.

## Bundle format
The bundle format is currently not fixed and will be most likely changed in future.

It is a `.tar` file and may contain the following files:
- `template/mimetype`: The mimetype of the input package
- `template/luaprog.lua`: The Lua program that has been generated for the main content element e.g. `content.xml` in ODF
- `template/init.lua: `The initialization program that runs before the `luaprog.lua` script
- `template/luaprog.nodelist`: The mapping between the regeneration commands in `loaprog.lua` to the main content element XML
- `template/content.xmltree`: The parsed XML tree of the main content XML in a human readable format
- `processed/content.xml`: The processed/generated main content XML after the Lua engine has run
- `processed/exec_trace.lua`: The execution trace while running the `luaprog.lua`, where the `processed/content.xml` is generated

# Concepts

## Rea

![rea placement](./rea.drawio.png)

## Templating pipeline
The templating process in rea works with the following scheme on a high level:

1. File decomposing: E.g. loading XMLs and mimetype from the ZIP/ODT in case of ODF
2. Parsing the needed XML data as tree: E.g. `content.xml`
3. Substituting, tokenizing and converting the XML tree to a Lua program that would regenerate the original XML but has control blocks in place
4. Executing the Lua program to get an execution trace
5. The execution trace will regenerate the XML while incorporating all control blocks. There is need for XML rebalancing
6. Document composing: E.g. repacking an ODT file
